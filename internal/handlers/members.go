package handlers

import (
	"fmt"
	"strings"

	"perimatrik-api-go/internal/export"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/request"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
)

func memberRules(uniqueNoAnggota string) []validation.Rule {
	return []validation.Rule{
		validation.R("no_anggota", "required|string|max:255|"+uniqueNoAnggota),
		validation.R("nama_lengkap", "required|string|max:255"),
		validation.R("nama_lapangan", "nullable|string|max:255"),
		validation.R("domisili", "nullable|string|max:255"),
		validation.R("kontak", "nullable|string|max:255"),
		validation.R("angkatan_id", "required|uuid|exists:angkatans,id"),
		validation.R("user_id", "nullable|uuid|exists:users,id"),
		validation.R("instagram", "nullable|string|max:255"),
		validation.R("facebook", "nullable|string|max:255"),
		validation.R("linkedin", "nullable|string|max:255"),
		validation.R("tiktok", "nullable|string|max:255"),
		validation.R("youtube", "nullable|string|max:255"),
	}
}

// fillableMember returns the present, mass-assignable member fields of in.
func fillableMember(in *request.Input) map[string]any {
	out := map[string]any{}
	for _, k := range models.MemberFillable {
		if k != "avatar" && in.Has(k) {
			out[k] = in.Str(k)
		}
	}
	return out
}

func (h *Handler) membersSorted() ([]models.Member, error) {
	members := []models.Member{}
	err := h.DB.Preload("Angkatan").
		Joins("JOIN angkatans ON members.angkatan_id = angkatans.id").
		Order("angkatans.year desc").
		Select("members.*").
		Find(&members).Error
	return members, err
}

// ListMembers — GET /content/member
func (h *Handler) ListMembers(c *gin.Context) {
	members, err := h.membersSorted()
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, members)
}

// ExportMembers — GET /content/export-members?format=xlsx|csv
func (h *Handler) ExportMembers(c *gin.Context) {
	members, err := h.membersSorted()
	if err != nil {
		h.serverError(c, err)
		return
	}
	rows := export.MemberRows(members)

	if strings.ToLower(c.DefaultQuery("format", "xlsx")) == "csv" {
		c.Header("Content-Disposition", `attachment; filename=database_anggota_perimatrik.csv`)
		c.Data(200, "text/csv; charset=UTF-8", export.MembersCSV(rows))
		return
	}
	data, err := export.MembersXLSX(rows)
	if err != nil {
		h.serverError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename=database_anggota_perimatrik.xlsx`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) findMember(id string) (*models.Member, error) {
	var m models.Member
	err := h.DB.Preload("Angkatan").First(&m, "id = ?", id).Error
	return &m, err
}

// ShowMember — GET /content/member/:id
func (h *Handler) ShowMember(c *gin.Context) {
	m, err := h.findMember(c.Param("id"))
	if err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	response.Data(c, 200, m)
}

// CreateMember — POST /content/member (admin)
func (h *Handler) CreateMember(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	rules := memberRules("unique:members")
	if in.HasFile("avatar") {
		rules = append(rules, validation.R("avatar", "required|"+imageRule))
	} else {
		rules = append(rules, validation.R("avatar", "nullable|string|url"))
	}
	if errs := h.V.Validate(in, rules...); !errs.Empty() {
		response.ValidationErrors(c, 422, response.MsgFailedCreate, errs)
		return
	}

	fields := fillableMember(in)
	fields["avatar"] = in.Str("avatar")
	if f := in.File("avatar"); f != nil {
		u, err := h.Storage.SaveURL(f, storage.MemberDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		fields["avatar"] = u
	}

	id := models.NewUUID()
	fields["id"] = id
	fields["created_at"] = models.Now()
	fields["updated_at"] = fields["created_at"]
	if err := h.DB.Table("members").Create(fields).Error; err != nil {
		h.serverError(c, err)
		return
	}
	m, err := h.findMember(id)
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 201, m)
}

// UpdateMember — PUT|POST /content/member/:id (admin)
func (h *Handler) UpdateMember(c *gin.Context) {
	m, err := h.findMember(c.Param("id"))
	if err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	in := h.input(c)
	if in == nil {
		return
	}

	data := in.Except("avatar", "_method")
	data.SetFile("avatar", in.File("avatar"))
	rules := append(memberRules(fmt.Sprintf("unique:members,no_anggota,%s", m.ID)),
		validation.R("avatar", "nullable|"+imageRule))
	if errs := h.V.Validate(data, rules...); !errs.Empty() {
		response.ValidationErrors(c, 422, response.MsgFailedUpdate, errs)
		return
	}

	fields := fillableMember(in)
	if f := in.File("avatar"); f != nil {
		if m.Avatar != nil && strings.Contains(*m.Avatar, "/storage/") {
			h.Storage.DeleteByURL(storage.MemberDir, *m.Avatar)
		}
		u, err := h.Storage.SaveURL(f, storage.MemberDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		fields["avatar"] = u
	}
	if len(fields) > 0 {
		if err := h.DB.Model(&models.Member{ID: m.ID}).Updates(fields).Error; err != nil {
			h.serverError(c, err)
			return
		}
	}
	m, err = h.findMember(m.ID)
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, m)
}

// DeleteMember — DELETE /content/member/:id (admin)
func (h *Handler) DeleteMember(c *gin.Context) {
	var m models.Member
	if err := h.DB.First(&m, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, "Member not found")
		} else {
			h.serverError(c, err)
		}
		return
	}
	if m.Avatar != nil && strings.Contains(*m.Avatar, "/storage/") {
		h.Storage.DeleteByURL(storage.MemberDir, *m.Avatar)
	}
	if err := h.DB.Delete(&m).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, "Member deleted successfully")
}

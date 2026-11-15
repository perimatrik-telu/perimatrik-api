package handlers

import (
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
)

// ListAngkatan — GET /content/angkatan
func (h *Handler) ListAngkatan(c *gin.Context) {
	items := []models.Angkatan{}
	if err := h.DB.Find(&items).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, items)
}

func (h *Handler) findAngkatan(c *gin.Context) *models.Angkatan {
	var a models.Angkatan
	if err := h.DB.First(&a, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, "Angkatan not found")
		} else {
			h.serverError(c, err)
		}
		return nil
	}
	return &a
}

// ShowAngkatan — GET /content/angkatan/:id
func (h *Handler) ShowAngkatan(c *gin.Context) {
	if a := h.findAngkatan(c); a != nil {
		response.Data(c, 200, a)
	}
}

// CreateAngkatan — POST /content/angkatan (admin)
func (h *Handler) CreateAngkatan(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("nama_angkatan", "required|string|max:255|unique:angkatans"),
		validation.R("year", "required|string|max:4"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 422, "Validation Error", errs)
		return
	}
	a := models.Angkatan{NamaAngkatan: in.String("nama_angkatan"), Year: in.Str("year")}
	if err := h.DB.Create(&a).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 201, a)
}

// UpdateAngkatan — PUT /content/angkatan/:id (admin)
func (h *Handler) UpdateAngkatan(c *gin.Context) {
	a := h.findAngkatan(c)
	if a == nil {
		return
	}
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("nama_angkatan", "required|string|max:255|unique:angkatans,nama_angkatan,"+a.ID),
		validation.R("year", "required|string|max:4"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 422, "Validation Error", errs)
		return
	}
	a.NamaAngkatan = in.String("nama_angkatan")
	a.Year = in.Str("year")
	if err := h.DB.Select("*").Omit("created_at").Save(a).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, a)
}

// DeleteAngkatan — DELETE /content/angkatan/:id (admin)
func (h *Handler) DeleteAngkatan(c *gin.Context) {
	a := h.findAngkatan(c)
	if a == nil {
		return
	}
	if err := h.DB.Delete(a).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, "Angkatan deleted successfully")
}

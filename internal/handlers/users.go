package handlers

import (
	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
)

var memberProfileFields = []string{
	"no_anggota", "nama_lapangan", "domisili", "kontak",
	"instagram", "facebook", "linkedin", "tiktok", "youtube",
}

func (h *Handler) findUserWithMember(id string) (*models.User, error) {
	var u models.User
	err := h.DB.Preload("Member.Angkatan").First(&u, "id = ?", id).Error
	return &u, err
}

// ListUsers — GET /system/users (admin)
func (h *Handler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Preload("Member.Angkatan").Order("created_at desc").Find(&users).Error; err != nil {
		h.serverError(c, err)
		return
	}
	out := make([]models.UserWithMember, len(users))
	for i, u := range users {
		out[i] = u.WithMember()
	}
	response.Data(c, 200, out)
}

// CreateUser — POST /system/users (admin)
func (h *Handler) CreateUser(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("name", "required|string|max:255"),
		validation.R("email", "required|string|email|max:255|unique:users"),
		validation.R("password", "required|string|min:8"),
		validation.R("role", "required|in:admin,officer,member"),
		validation.R("avatar", "nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 422, response.MsgFailedCreate, errs)
		return
	}

	var avatar *string
	if f := in.File("avatar"); f != nil {
		u, err := h.Storage.SaveURL(f, storage.AvatarDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		avatar = &u
	}
	hash, err := h.Auth.HashPassword(in.String("password"))
	if err != nil {
		h.serverError(c, err)
		return
	}

	user := models.User{
		Name: in.String("name"), Email: in.String("email"), Password: hash,
		Role: in.String("role"), Avatar: avatar,
	}
	if err := h.DB.Create(&user).Error; err != nil {
		h.serverError(c, err)
		return
	}

	if in.Filled("member_id") {
		var m models.Member
		if h.DB.First(&m, "id = ?", in.String("member_id")).Error == nil {
			h.DB.Model(&m).Updates(map[string]any{
				"user_id": user.ID, "nama_lengkap": user.Name, "avatar": user.Avatar,
			})
		}
	}

	response.Data(c, 201, user)
}

// ShowUser — GET /system/users/:id and /profile/:id
func (h *Handler) ShowUser(c *gin.Context) {
	me, id := auth.MustUser(c), c.Param("id")
	if !isAdmin(me) && me.ID != id {
		response.Message(c, 403, false, "Forbidden. You can only view your own profile.")
		return
	}
	user, err := h.findUserWithMember(id)
	if err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	response.Data(c, 200, user.WithMember())
}

// UpdateUser — PUT|POST /system/users/:id and /profile/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	me, id := auth.MustUser(c), c.Param("id")
	if !isAdmin(me) && me.ID != id {
		response.Message(c, 403, false, "Forbidden. You can only update your own profile.")
		return
	}

	var user models.User
	if err := h.DB.First(&user, "id = ?", id).Error; err != nil {
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
	rules := []validation.Rule{
		validation.R("name", "required|string|max:255"),
		validation.R("email", "required|string|email|max:255|unique:users,email,"+user.ID),
		validation.R("password", "nullable|string|min:8"),
		validation.R("avatar", "nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048"),
	}
	for _, f := range memberProfileFields {
		rules = append(rules, validation.R(f, "nullable|string|max:255"))
	}
	if isAdmin(me) {
		rules = append(rules, validation.R("role", "required|in:admin,officer,member"))
	}
	if errs := h.V.Validate(in, rules...); !errs.Empty() {
		response.ValidationErrors(c, 422, response.MsgFailedUpdate, errs)
		return
	}

	updates := map[string]any{"name": in.String("name"), "email": in.String("email")}
	if f := in.File("avatar"); f != nil {
		if user.Avatar != nil {
			h.Storage.DeleteByURL(storage.AvatarDir, *user.Avatar)
		}
		u, err := h.Storage.SaveURL(f, storage.AvatarDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		user.Avatar = &u
		updates["avatar"] = u
	}
	if isAdmin(me) && in.Filled("role") {
		updates["role"] = in.String("role")
	}
	if in.Filled("password") {
		hash, err := h.Auth.HashPassword(in.String("password"))
		if err != nil {
			h.serverError(c, err)
			return
		}
		updates["password"] = hash
	}
	if err := h.DB.Model(&user).Updates(updates).Error; err != nil {
		h.serverError(c, err)
		return
	}

	// Sync with the linked member record.
	var member models.Member
	if h.DB.Where("user_id = ?", user.ID).First(&member).Error == nil {
		sync := map[string]any{"nama_lengkap": in.String("name"), "avatar": user.Avatar}
		for _, f := range memberProfileFields {
			if in.Has(f) {
				if f == "no_anggota" && !in.Filled(f) {
					continue // NOT NULL column
				}
				sync[f] = in.Str(f)
			}
		}
		if err := h.DB.Model(&member).Updates(sync).Error; err != nil {
			h.serverError(c, err)
			return
		}
	}

	fresh, err := h.findUserWithMember(user.ID)
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, fresh.WithMember())
}

// DeleteUser — DELETE /system/users/:id (admin)
func (h *Handler) DeleteUser(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	if user.ID == auth.MustUser(c).ID {
		response.Message(c, 400, false, "You cannot delete your own account.")
		return
	}
	if err := h.DB.Delete(&user).Error; err != nil {
		h.serverError(c, err)
		return
	}
	h.DB.Where("tokenable_type = ? AND tokenable_id = ?", models.UserMorphType, user.ID).
		Delete(&models.PersonalAccessToken{})
	response.Message(c, 200, true, "User deleted successfully.")
}

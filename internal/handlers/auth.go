package handlers

import (
	"strings"

	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/support"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register — POST /auth/register
func (h *Handler) Register(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}

	if !support.FilterBool(h.settingValue("allow_member_registration", "1")) {
		validationException(c, validation.Single("register", "Registrasi anggota sedang dinonaktifkan oleh admin."))
		return
	}

	errs := h.V.Validate(in,
		validation.R("no_anggota", "required|string|max:255"),
		validation.R("nama_anggota", "required|string|max:255"),
		validation.R("email", "required|string|email|max:255|unique:users"),
		validation.R("password", "required|string|min:8|confirmed"),
	)
	if !errs.Empty() {
		validationException(c, errs)
		return
	}

	var member models.Member
	if err := h.DB.Where("no_anggota = ?", in.String("no_anggota")).First(&member).Error; err != nil {
		if notFound(err) {
			validationException(c, validation.Single("no_anggota", "Nomor anggota tidak ditemukan."))
		} else {
			h.serverError(c, err)
		}
		return
	}
	if !strings.EqualFold(strings.TrimSpace(member.NamaLengkap), strings.TrimSpace(in.String("nama_anggota"))) {
		validationException(c, validation.Single("nama_anggota", "Nama anggota tidak sesuai dengan nomor anggota."))
		return
	}
	if member.UserID != nil {
		validationException(c, validation.Single("no_anggota", "Nomor anggota ini sudah terhubung ke akun pengguna."))
		return
	}

	hash, err := h.Auth.HashPassword(in.String("password"))
	if err != nil {
		h.serverError(c, err)
		return
	}
	user := models.User{Name: member.NamaLengkap, Email: in.String("email"), Password: hash, Role: "member"}
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Model(&member).Update("user_id", user.ID).Error
	})
	if err != nil {
		h.serverError(c, err)
		return
	}

	response.Success(c, 201, "Registrasi berhasil. Silakan login menggunakan email dan password Anda.", user)
}

// Login — POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("email", "required|email"),
		validation.R("password", "required"),
	)
	if !errs.Empty() {
		validationException(c, errs)
		return
	}

	var user models.User
	err := h.DB.Where("email = ?", in.String("email")).First(&user).Error
	if err != nil && !notFound(err) {
		h.serverError(c, err)
		return
	}
	if err != nil || !auth.CheckPassword(in.String("password"), user.Password) {
		const msg = "The provided credentials are incorrect."
		response.Error(c, 422, msg, validation.Single("email", msg), response.CodeInvalidCredentials)
		return
	}

	token, err := h.Auth.CreateToken(user.ID, "auth_token")
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Success(c, 200, response.MsgLoginSuccess, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}

// CurrentUser — GET /auth/current-user and GET /user
func (h *Handler) CurrentUser(c *gin.Context) {
	response.Success(c, 200, response.MsgFound, auth.MustUser(c))
}

// Logout — POST /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	if t := auth.Token(c); t != nil {
		if err := h.DB.Delete(t).Error; err != nil {
			h.serverError(c, err)
			return
		}
	}
	response.Success(c, 200, response.MsgLogoutSuccess, nil)
}

// ChangePassword — POST /auth/change-password
func (h *Handler) ChangePassword(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("old_password", "required|string"),
		validation.R("password", "required|string|min:8|confirmed"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 422, "Validasi gagal.", errs)
		return
	}

	user := auth.MustUser(c)
	if !auth.CheckPassword(in.String("old_password"), user.Password) {
		response.Message(c, 400, false, "Password lama tidak sesuai.")
		return
	}
	hash, err := h.Auth.HashPassword(in.String("password"))
	if err != nil {
		h.serverError(c, err)
		return
	}
	if err := h.DB.Model(user).Update("password", hash).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, "Password berhasil diubah.")
}

// VerifyPassword — POST /auth/verify-password
func (h *Handler) VerifyPassword(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in, validation.R("old_password", "required|string"))
	if !errs.Empty() {
		response.ValidationErrors(c, 422, "Validasi gagal.", errs)
		return
	}
	if !auth.CheckPassword(in.String("old_password"), auth.MustUser(c).Password) {
		response.Message(c, 400, false, "Password lama tidak sesuai.")
		return
	}
	response.Message(c, 200, true, "Password lama sesuai.")
}

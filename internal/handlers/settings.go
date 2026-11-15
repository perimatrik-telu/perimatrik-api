package handlers

import (
	"encoding/json"

	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/request"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/support"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var contactKeys = []string{"contact_email", "contact_phone", "contact_address", "contact_maps", "allow_member_registration"}

// setting returns the stored value (nil if stored as NULL) or def when the
// key does not exist.
func (h *Handler) setting(key, def string) *string {
	var s models.Setting
	if err := h.DB.Where("`key` = ?", key).First(&s).Error; err != nil {
		return &def
	}
	return s.Value
}

func (h *Handler) settingValue(key, def string) string {
	if v := h.setting(key, def); v != nil {
		return *v
	}
	return ""
}

// GetContactInfo — GET /system/settings/contact
func (h *Handler) GetContactInfo(c *gin.Context) {
	response.Data(c, 200, gin.H{
		"contact_email":             h.setting("contact_email", ""),
		"contact_phone":             h.setting("contact_phone", ""),
		"contact_address":           h.setting("contact_address", ""),
		"contact_maps":              h.setting("contact_maps", ""),
		"allow_member_registration": support.FilterBool(h.settingValue("allow_member_registration", "1")),
	})
}

// UpdateContactInfo — POST /system/settings/contact (admin)
func (h *Handler) UpdateContactInfo(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("contact_email", "nullable|email|max:255"),
		validation.R("contact_phone", "nullable|string|max:255"),
		validation.R("contact_address", "nullable|string"),
		validation.R("contact_maps", "nullable|string"),
		validation.R("allow_member_registration", "required|boolean"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 422, "Validation failed", errs)
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, key := range contactKeys {
			raw, ok := in.Get(key)
			if !ok {
				continue
			}
			var value *string
			switch v := raw.(type) {
			case nil:
			case bool:
				value = models.Str(map[bool]string{true: "1", false: "0"}[v])
			case json.Number, string:
				value = models.Str(request.Stringify(v))
			}
			s := models.Setting{ID: models.NewUUID(), Key: key, Value: value}
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
			}).Create(&s).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, "Contact information updated successfully.")
}

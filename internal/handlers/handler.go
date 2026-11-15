// Package handlers contains the HTTP handlers, one file per Laravel controller.
package handlers

import (
	"errors"
	"log"
	"time"

	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/config"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/request"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/support"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB      *gorm.DB
	Cfg     *config.Config
	Auth    *auth.Service
	Storage *storage.Storage
	V       *validation.Validator
}

// input parses the request input; on failure it writes a 400 and returns nil.
func (h *Handler) input(c *gin.Context) *request.Input {
	in, err := request.Parse(c.Request)
	if err != nil {
		response.Error(c, 400, err.Error(), nil, 0)
		return nil
	}
	return in
}

// serverError logs err and writes a 500, exposing the message only in debug.
func (h *Handler) serverError(c *gin.Context, err error) {
	log.Printf("ERROR %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	msg := response.MsgInternalServerError
	if h.Cfg.AppDebug {
		msg = err.Error()
	}
	response.Error(c, 500, msg, nil, 0)
}

// validationException mirrors a thrown ValidationException (422 + summary).
func validationException(c *gin.Context, errs *validation.Errors) {
	response.Error(c, 422, errs.Summary(), errs, response.CodeValidationFailed)
}

func notFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// now returns the current time in the app timezone.
func (h *Handler) now() time.Time { return time.Now().In(h.Cfg.Location) }

// parseDate parses a request date in the app timezone.
func (h *Handler) parseDate(s string) *models.Time {
	if t, ok := support.ParseDate(s, h.Cfg.Location); ok {
		return models.TimePtr(t)
	}
	return nil
}

// pagination reads page/limit like the Laravel controllers (defaults 1/10).
func pagination(in *request.Input) (page, limit, offset int) {
	page = max(in.Int("page", 1), 1)
	limit = max(in.Int("limit", 10), 1)
	return page, limit, (page - 1) * limit
}

// sortDirection whitelists an ORDER BY direction (default DESC).
func sortDirection(s string) string {
	if s == "asc" || s == "ASC" || s == "Asc" {
		return "ASC"
	}
	return "DESC"
}

func isAdmin(u *models.User) bool { return u != nil && u.Role == "admin" }

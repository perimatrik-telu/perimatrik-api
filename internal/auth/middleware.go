package auth

import (
	"slices"
	"strings"

	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"

	"github.com/gin-gonic/gin"
)

const (
	ctxUser     = "auth.user"
	ctxToken    = "auth.token"
	ctxResolved = "auth.resolved"
)

// User resolves (once per request) the user of the bearer token, or nil.
// Equivalent to Auth::guard('sanctum')->user().
func (s *Service) User(c *gin.Context) *models.User {
	if _, done := c.Get(ctxResolved); !done {
		c.Set(ctxResolved, true)
		if u, t, err := s.Authenticate(bearerToken(c)); err == nil {
			c.Set(ctxUser, u)
			c.Set(ctxToken, t)
		}
	}
	if u, ok := c.Get(ctxUser); ok {
		return u.(*models.User)
	}
	return nil
}

// Token returns the access token used for this request (after User()).
func Token(c *gin.Context) *models.PersonalAccessToken {
	if t, ok := c.Get(ctxToken); ok {
		return t.(*models.PersonalAccessToken)
	}
	return nil
}

// MustUser returns the authenticated user; only call behind Required().
func MustUser(c *gin.Context) *models.User {
	return c.MustGet(ctxUser).(*models.User)
}

// Required is the auth:sanctum middleware.
func (s *Service) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.User(c) == nil {
			response.Error(c, 401, "Unauthenticated.", nil, 0)
			return
		}
		c.Next()
	}
}

// Role is the role:a,b middleware.
func (s *Service) Role(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := s.User(c)
		if u == nil {
			response.Error(c, 401, "Unauthenticated.", nil, 0)
			return
		}
		if !slices.Contains(roles, u.Role) {
			response.Error(c, 403, "Forbidden. You do not have the required role.", nil, 0)
			return
		}
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

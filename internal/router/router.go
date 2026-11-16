package router

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"perimatrik-api-go/internal/config"
	"perimatrik-api-go/internal/handlers"
	"perimatrik-api-go/internal/response"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, h *handlers.Handler) *gin.Engine {
	if !cfg.AppDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.MaxMultipartMemory = 32 << 20
	_ = r.SetTrustedProxies(nil)

	r.Use(gin.Logger(), recovery(cfg), cors(cfg))

	r.NoRoute(func(c *gin.Context) {
		if isAPI(c) {
			response.Error(c, 404, response.MsgNoItem, nil, 0)
			return
		}
		c.String(404, "Not Found")
	})
	r.NoMethod(func(c *gin.Context) {
		msg := fmt.Sprintf("The %s method is not supported for route %s.", c.Request.Method, strings.TrimPrefix(c.Request.URL.Path, "/"))
		response.Error(c, 405, msg, nil, 0)
	})

	// --- Web routes ---
	r.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Hello World"}) })
	r.GET("/up", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/share/article/:slug", h.ShareArticle)
	r.GET("/share/news/:slug", h.ShareNews)
	r.GET("/storage/*path", h.ShowImage)
	r.HEAD("/storage/*path", h.ShowImage)

	a := h.Auth
	authed := a.Required()
	api := r.Group("/api/v1")

	// --- Authentication ---
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/logout", authed, h.Logout)
	api.POST("/auth/change-password", authed, h.ChangePassword)
	api.POST("/auth/verify-password", authed, h.VerifyPassword)
	api.GET("/auth/current-user", authed, h.CurrentUser)
	api.GET("/user", authed, h.CurrentUser)

	// --- User management ---
	admin := a.Role("admin")
	api.GET("/system/users", authed, admin, h.ListUsers)
	api.POST("/system/users", authed, admin, h.CreateUser)
	api.DELETE("/system/users/:id", authed, admin, h.DeleteUser)
	for _, p := range []string{"/system/users/:id", "/profile/:id"} {
		api.GET(p, authed, h.ShowUser)
		api.PUT(p, authed, h.UpdateUser)
		api.POST(p, authed, h.UpdateUser) // FormData
	}

	// --- Settings ---
	api.GET("/system/settings/contact", h.GetContactInfo)
	api.POST("/system/settings/contact", authed, admin, h.UpdateContactInfo)

	// --- Articles ---
	api.GET("/content/article", h.ListArticles)
	api.GET("/content/article/:slug", h.ShowArticle)
	api.GET("/content/article/id/:id", h.ShowArticleByID)
	api.GET("/content/article/:slug/related", h.RelatedArticles)
	writer := a.Role("admin", "member")
	api.POST("/content/article", authed, writer, h.CreateArticle)
	api.POST("/content/article/:id", authed, writer, h.UpdateArticle)
	api.PUT("/content/article/:id", authed, writer, h.UpdateArticle)
	api.DELETE("/content/article/:id", authed, writer, h.DeleteArticle)

	// --- News ---
	api.GET("/content/news", h.ListNews)
	api.GET("/content/news/:slug", h.ShowNews)
	api.GET("/content/news/id/:id", h.ShowNewsByID)
	api.GET("/content/news/:slug/related", h.RelatedNews)
	officer := a.Role("admin", "officer")
	api.POST("/content/news", authed, officer, h.CreateNews)
	api.POST("/content/news/:id", authed, officer, h.UpdateNews)
	api.PUT("/content/news/:id", authed, officer, h.UpdateNews)
	api.DELETE("/content/news/:id", authed, officer, h.DeleteNews)

	// --- Gallery ---
	api.GET("/content/gallery", h.ListGallery)
	api.GET("/content/gallery/:id", h.ShowGallery)
	api.GET("/images/*path", h.ShowImage)
	anyRole := a.Role("admin", "officer", "member")
	api.POST("/content/gallery", authed, anyRole, h.CreateGallery)
	api.POST("/content/gallery/:id", authed, anyRole, h.UpdateGallery)
	api.PUT("/content/gallery/:id", authed, anyRole, h.UpdateGallery)
	api.DELETE("/content/gallery/:id", authed, anyRole, h.DeleteGallery)

	// --- Members ---
	api.GET("/content/member", h.ListMembers)
	api.GET("/content/export-members", h.ExportMembers)
	api.GET("/content/member/:id", h.ShowMember)
	api.POST("/content/member", authed, admin, h.CreateMember)
	api.POST("/content/member/:id", authed, admin, h.UpdateMember)
	api.PUT("/content/member/:id", authed, admin, h.UpdateMember)
	api.DELETE("/content/member/:id", authed, admin, h.DeleteMember)

	// --- Angkatan ---
	api.GET("/content/angkatan", h.ListAngkatan)
	api.GET("/content/angkatan/:id", h.ShowAngkatan)
	api.POST("/content/angkatan", authed, admin, h.CreateAngkatan)
	api.PUT("/content/angkatan/:id", authed, admin, h.UpdateAngkatan)
	api.DELETE("/content/angkatan/:id", authed, admin, h.DeleteAngkatan)

	// --- Comments ---
	api.GET("/content/comments", h.ListComments)
	api.POST("/content/comments", authed, h.CreateComment)
	api.DELETE("/content/comments/:id", authed, h.DeleteComment)

	// --- Status ---
	api.GET("/admin-only", authed, admin, func(c *gin.Context) {
		response.Success(c, 200, "Welcome, Admin!", nil)
	})

	return r
}

func isAPI(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/") ||
		strings.Contains(c.GetHeader("Accept"), "json")
}

func recovery(cfg *config.Config) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("PANIC %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		msg := response.MsgInternalServerError
		if cfg.AppDebug {
			msg = fmt.Sprint(err)
		}
		response.Error(c, http.StatusInternalServerError, msg, nil, 0)
	})
}

// cors mirrors config/cors.php: listed origins + patterns, any method/header,
// credentials allowed.
func cors(cfg *config.Config) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range cfg.CORSOrigins {
		allowed[o] = true
	}
	originOK := func(o string) bool {
		if allowed[o] {
			return true
		}
		for _, re := range cfg.CORSOriginPatterns {
			if re.MatchString(o) {
				return true
			}
		}
		return false
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		preflight := c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != ""
		h := c.Writer.Header()
		h.Add("Vary", "Origin")

		if origin != "" && originOK(origin) {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			if preflight {
				h.Set("Access-Control-Allow-Methods", strings.ToUpper(c.GetHeader("Access-Control-Request-Method")))
				if rh := c.GetHeader("Access-Control-Request-Headers"); rh != "" {
					h.Set("Access-Control-Allow-Headers", rh)
				}
			}
		}
		if preflight {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

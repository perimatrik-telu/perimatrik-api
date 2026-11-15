package handlers

import (
	"embed"
	"html/template"
	"strings"

	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/support"

	"github.com/gin-gonic/gin"
)

//go:embed templates/share.html
var templatesFS embed.FS

var shareTmpl = template.Must(template.ParseFS(templatesFS, "templates/share.html"))

type sharePage struct {
	Title, Description, Image, URL string
}

// ShareArticle — GET /share/article/:slug (Open Graph page that redirects to the frontend)
func (h *Handler) ShareArticle(c *gin.Context) {
	var a models.Article
	if err := h.DB.Where("slug = ?", c.Param("slug")).First(&a).Error; err != nil {
		shareNotFound(c)
		return
	}
	h.renderShare(c, a.Title, a.Body, a.Image, "/article/detail/"+a.Slug)
}

// ShareNews — GET /share/news/:slug
func (h *Handler) ShareNews(c *gin.Context) {
	var n models.News
	if err := h.DB.Where("slug = ?", c.Param("slug")).First(&n).Error; err != nil {
		shareNotFound(c)
		return
	}
	image := ""
	if n.Image != nil {
		image = *n.Image
	}
	h.renderShare(c, n.Title, n.Body, image, "/news/detail/"+n.Slug)
}

func (h *Handler) renderShare(c *gin.Context, title, body, image, path string) {
	c.Header("Content-Type", "text/html; charset=UTF-8")
	c.Status(200)
	_ = shareTmpl.Execute(c.Writer, sharePage{
		Title:       title,
		Description: support.Limit(support.StripTags(body), 150),
		Image:       image,
		URL:         strings.TrimRight(h.Cfg.FrontendURL, "/") + path,
	})
}

func shareNotFound(c *gin.Context) {
	if strings.Contains(c.GetHeader("Accept"), "json") {
		response.Error(c, 404, response.MsgNoItem, nil, 0)
		return
	}
	c.String(404, "Not Found")
}

package handlers

import (
	"strings"

	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListNews — GET /content/news
func (h *Handler) ListNews(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	page, limit, offset := pagination(in)

	q := h.DB.Model(&models.News{})
	if uid := in.String("user_id"); uid != "" && uid != "undefined" && uid != "0" {
		q = q.Where("user_id = ?", uid)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		h.serverError(c, err)
		return
	}
	news := []models.News{}
	err := q.Preload("User").Order("published_at " + sortDirection(in.String("filter"))).
		Offset(offset).Limit(limit).Find(&news).Error
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Paginated(c, news, page, limit, total)
}

// ShowNews — GET /content/news/:slug
func (h *Handler) ShowNews(c *gin.Context) { h.showNews(c, h.DB.Where("slug = ?", c.Param("slug"))) }

// ShowNewsByID — GET /content/news/id/:id
func (h *Handler) ShowNewsByID(c *gin.Context) { h.showNews(c, h.DB.Where("id = ?", c.Param("id"))) }

func (h *Handler) showNews(c *gin.Context, q *gorm.DB) {
	var n models.News
	if err := q.Preload("User").First(&n).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	if err := h.incrementViews(&n, &n.Views, &n.UpdatedAt); err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, n)
}

// RelatedNews — GET /content/news/:slug/related
func (h *Handler) RelatedNews(c *gin.Context) {
	news := []models.News{}
	err := h.DB.Preload("User").Where("slug <> ?", c.Param("slug")).
		Order("created_at desc").Limit(3).Find(&news).Error
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, news)
}

// CreateNews — POST /content/news (admin, officer)
func (h *Handler) CreateNews(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	rules := []validation.Rule{
		validation.R("title", "required|string"),
		validation.R("body", "required|string"),
		validation.R("published_at", "nullable|date"),
	}
	if in.HasFile("image") {
		rules = append(rules, validation.R("image", "required|"+imageRule))
	} else {
		rules = append(rules, validation.R("image", "required|string|url"))
	}
	if errs := h.V.Validate(in, rules...); !errs.Empty() {
		response.ValidationErrors(c, 400, response.MsgFailedCreate, errs)
		return
	}

	image := in.String("image")
	if f := in.File("image"); f != nil {
		u, err := h.Storage.SaveURL(f, storage.NewsDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		image = u
	}

	n := models.News{
		Slug:        makeSlug(in.String("title")),
		Image:       &image,
		Title:       in.String("title"),
		Body:        in.String("body"),
		UserID:      &auth.MustUser(c).ID,
		PublishedAt: models.TimePtr(h.now()),
	}
	if p := in.Str("published_at"); p != nil {
		n.PublishedAt = h.parseDate(*p)
	}
	if err := h.DB.Create(&n).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 201, true, response.MsgSuccessCreate)
}

// UpdateNews — PUT|POST /content/news/:id (admin, officer)
func (h *Handler) UpdateNews(c *gin.Context) {
	var n models.News
	if err := h.DB.First(&n, "id = ?", c.Param("id")).Error; err != nil {
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

	data := in.Only("title", "body", "published_at")
	data.SetFile("image", in.File("image"))
	errs := h.V.Validate(data,
		validation.R("title", "required|string"),
		validation.R("body", "required|string"),
		validation.R("image", "nullable|"+imageRule),
		validation.R("published_at", "nullable|date"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 400, response.MsgFailedUpdate, errs)
		return
	}

	if f := in.File("image"); f != nil {
		if n.Image != nil && strings.Contains(*n.Image, "/storage/") {
			h.Storage.DeleteByURL(storage.NewsDir, *n.Image)
		}
		u, err := h.Storage.SaveURL(f, storage.NewsDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		n.Image = &u
	}

	title := in.String("title")
	if title != n.Title {
		n.Slug = makeSlug(title)
	}
	n.Title = title
	n.Body = in.String("body")
	if in.Has("published_at") {
		n.PublishedAt = nil
		if p := in.Str("published_at"); p != nil {
			n.PublishedAt = h.parseDate(*p)
		}
	}

	if err := h.DB.Select("*").Omit("created_at").Save(&n).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, response.MsgSuccessUpdate)
}

// DeleteNews — DELETE /content/news/:id (admin, officer)
func (h *Handler) DeleteNews(c *gin.Context) {
	var n models.News
	if err := h.DB.First(&n, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	if n.Image != nil && strings.Contains(*n.Image, "/storage/") {
		h.Storage.DeleteByURL(storage.NewsDir, *n.Image)
	}
	if err := h.DB.Delete(&n).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, response.MsgSuccessDelete)
}

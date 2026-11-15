package handlers

import (
	"fmt"
	"strings"
	"time"

	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/support"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const imageRule = "file|mimes:jpeg,png,jpg,gif,svg|max:2048"

func makeSlug(title string) string {
	return fmt.Sprintf("%s-%d", support.Slug(title), time.Now().Unix())
}

// ListArticles — GET /content/article
func (h *Handler) ListArticles(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	page, limit, offset := pagination(in)
	userID := in.String("user_id")
	me := h.Auth.User(c)

	q := h.DB.Model(&models.Article{})
	if userID != "" && userID != "undefined" && userID != "0" {
		if me == nil || (!isAdmin(me) && me.ID != userID) {
			response.Message(c, 403, false, response.MsgUnauthorized)
			return
		}
		q = q.Where("user_id = ?", userID)
	} else if !isAdmin(me) {
		q = q.Where("status = ? AND published_at IS NOT NULL AND published_at <= ?", "active", h.now())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		h.serverError(c, err)
		return
	}
	articles := []models.Article{}
	err := q.Preload("User").Order("published_at " + sortDirection(in.String("filter"))).
		Offset(offset).Limit(limit).Find(&articles).Error
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Paginated(c, articles, page, limit, total)
}

// ShowArticle — GET /content/article/:slug
func (h *Handler) ShowArticle(c *gin.Context) {
	h.showArticle(c, h.DB.Where("slug = ?", c.Param("slug")))
}

// ShowArticleByID — GET /content/article/id/:id
func (h *Handler) ShowArticleByID(c *gin.Context) {
	h.showArticle(c, h.DB.Where("id = ?", c.Param("id")))
}

func (h *Handler) showArticle(c *gin.Context, q *gorm.DB) {
	var a models.Article
	if err := q.Preload("User").First(&a).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	if !h.canViewArticle(c, &a) {
		response.Message(c, 404, false, response.MsgNoItem)
		return
	}
	if err := h.incrementViews(&a, &a.Views, &a.UpdatedAt); err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, a)
}

// incrementViews mirrors $model->increment('views'): updated_at is touched in
// the database, but (like Eloquent) the returned model keeps the old value.
func (h *Handler) incrementViews(model any, views *uint, updatedAt *models.Time) error {
	before, count := *updatedAt, *views
	err := h.DB.Model(model).UpdateColumns(map[string]any{
		"views":      gorm.Expr("views + 1"),
		"updated_at": models.Now(),
	}).Error
	// GORM writes assigned columns back into the struct; restore them.
	*updatedAt, *views = before, count
	if err == nil {
		*views++
	}
	return err
}

func (h *Handler) canViewArticle(c *gin.Context, a *models.Article) bool {
	if a.Status == "active" && a.PublishedAt != nil && !a.PublishedAt.Std().After(time.Now()) {
		return true
	}
	me := h.Auth.User(c)
	return me != nil && (isAdmin(me) || (a.UserID != nil && *a.UserID == me.ID))
}

// RelatedArticles — GET /content/article/:slug/related
func (h *Handler) RelatedArticles(c *gin.Context) {
	articles := []models.Article{}
	err := h.DB.Preload("User").
		Where("slug <> ?", c.Param("slug")).
		Where("status = ? AND published_at IS NOT NULL AND published_at <= ?", "active", h.now()).
		Order("created_at desc").Limit(3).Find(&articles).Error
	if err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, articles)
}

// CreateArticle — POST /content/article (admin, member)
func (h *Handler) CreateArticle(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	isDraft := in.String("status") == "draft"
	req := map[bool]string{true: "nullable", false: "required"}[isDraft]

	rules := []validation.Rule{
		validation.R("title", req+"|string"),
		validation.R("body", req+"|string"),
		validation.R("published_at", "nullable|date"),
		validation.R("status", "nullable|in:draft,active,inactive"),
	}
	if in.HasFile("image") {
		rules = append(rules, validation.R("image", "nullable|"+imageRule))
	} else if isDraft {
		rules = append(rules, validation.R("image", "nullable|string"))
	} else {
		rules = append(rules, validation.R("image", "required|string|url"))
	}
	if errs := h.V.Validate(in, rules...); !errs.Empty() {
		response.ValidationErrors(c, 400, response.MsgFailedCreate, errs)
		return
	}

	image := in.String("image")
	if f := in.File("image"); f != nil {
		u, err := h.Storage.SaveURL(f, storage.ArticleDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		image = u
	}

	title := in.StringOr("title", "Untitled Draft")
	a := models.Article{
		Slug:   makeSlug(title),
		Image:  image,
		Title:  title,
		Body:   in.String("body"),
		UserID: &auth.MustUser(c).ID,
		Status: in.StringOr("status", "active"),
	}
	if a.Status == "active" {
		a.PublishedAt = models.TimePtr(h.now())
		if p := in.Str("published_at"); p != nil {
			a.PublishedAt = h.parseDate(*p)
		}
	}
	if err := h.DB.Create(&a).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 201, true, response.MsgSuccessCreate)
}

// UpdateArticle — PUT|POST /content/article/:id (admin, member)
func (h *Handler) UpdateArticle(c *gin.Context) {
	var a models.Article
	if err := h.DB.First(&a, "id = ?", c.Param("id")).Error; err != nil {
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

	isDraft := in.String("status") == "draft" || (a.Status == "draft" && !in.Has("status"))
	req := map[bool]string{true: "nullable", false: "required"}[isDraft]

	data := in.Only("title", "body", "published_at", "status")
	data.SetFile("image", in.File("image"))
	errs := h.V.Validate(data,
		validation.R("title", req+"|string"),
		validation.R("body", req+"|string"),
		validation.R("image", "nullable|"+imageRule),
		validation.R("published_at", "nullable|date"),
		validation.R("status", "nullable|in:draft,active,inactive"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 400, response.MsgFailedUpdate, errs)
		return
	}

	if f := in.File("image"); f != nil {
		if strings.Contains(a.Image, "/storage/") {
			h.Storage.DeleteByURL(storage.ArticleDir, a.Image)
		}
		u, err := h.Storage.SaveURL(f, storage.ArticleDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		a.Image = u
	}

	newTitle := in.StringOr("title", "Untitled Draft")
	if in.Has("title") {
		if newTitle != a.Title {
			a.Slug = makeSlug(newTitle)
		}
		a.Title = newTitle
	}
	if in.Has("body") || isDraft {
		a.Body = in.String("body")
	}
	if s := in.Str("status"); s != nil {
		a.Status = *s
	}

	switch {
	case a.Status == "draft":
		a.PublishedAt = nil
	case in.Has("published_at"):
		a.PublishedAt = models.TimePtr(h.now())
		if p := in.Str("published_at"); p != nil {
			a.PublishedAt = h.parseDate(*p)
		}
	case a.PublishedAt == nil:
		a.PublishedAt = models.TimePtr(h.now())
	}

	if err := h.DB.Select("*").Omit("created_at").Save(&a).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, response.MsgSuccessUpdate)
}

// DeleteArticle — DELETE /content/article/:id (admin, member)
func (h *Handler) DeleteArticle(c *gin.Context) {
	var a models.Article
	if err := h.DB.First(&a, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	if strings.Contains(a.Image, "/storage/") {
		h.Storage.DeleteByURL(storage.ArticleDir, a.Image)
	}
	if err := h.DB.Delete(&a).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, response.MsgSuccessDelete)
}

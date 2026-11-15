package handlers

import (
	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/storage"
	"perimatrik-api-go/internal/support"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
)

func gallerySlug(caption *string) *string {
	if caption != nil {
		return models.Str(support.Slug(*caption))
	}
	return models.Str(support.RandomString(10))
}

// ListGallery — GET /content/gallery
func (h *Handler) ListGallery(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	page, limit, offset := pagination(in)

	q := h.DB.Model(&models.Gallery{})
	if uid := in.String("user_id"); uid != "" && uid != "0" {
		q = q.Where("user_id = ?", uid)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		h.serverError(c, err)
		return
	}
	items := []models.Gallery{}
	if err := q.Preload("User").Order("updated_at desc").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Paginated(c, items, page, limit, total)
}

// ShowGallery — GET /content/gallery/:id (data is {} when not found, as before)
func (h *Handler) ShowGallery(c *gin.Context) {
	var g models.Gallery
	err := h.DB.Preload("User").First(&g, "id = ?", c.Param("id")).Error
	if err != nil {
		if !notFound(err) {
			h.serverError(c, err)
			return
		}
		response.Data(c, 200, nil)
		return
	}
	response.Data(c, 200, g)
}

var galleryRules = []validation.Rule{
	validation.R("caption", "nullable|string"),
	validation.R("latitude", "nullable|numeric|between:-90,90"),
	validation.R("longitude", "nullable|numeric|between:-180,180"),
}

// CreateGallery — POST /content/gallery (admin, officer, member)
func (h *Handler) CreateGallery(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	rules := append([]validation.Rule{validation.R("image", "required|file|mimes:jpeg,png,jpg,gif,svg|max:2024")}, galleryRules...)
	if errs := h.V.Validate(in, rules...); !errs.Empty() {
		response.ValidationErrors(c, 400, response.MsgFailedCreate, errs)
		return
	}

	url, err := h.Storage.SaveURL(in.File("image"), storage.GalleryDir)
	if err != nil {
		h.serverError(c, err)
		return
	}
	g := models.Gallery{
		Image:     url,
		Caption:   in.Str("caption"),
		Latitude:  in.Str("latitude"),
		Longitude: in.Str("longitude"),
		UserID:    &auth.MustUser(c).ID,
	}
	g.Slug = gallerySlug(g.Caption)
	if err := h.DB.Create(&g).Error; err != nil {
		h.serverError(c, err)
		return
	}
	h.attachUser(&g)
	response.Success(c, 201, response.MsgSuccessCreate, g)
}

// UpdateGallery — PUT|POST /content/gallery/:id (admin, officer, member)
func (h *Handler) UpdateGallery(c *gin.Context) {
	var g models.Gallery
	if err := h.DB.First(&g, "id = ?", c.Param("id")).Error; err != nil {
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
	rules := append([]validation.Rule{validation.R("image", "nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2024")}, galleryRules...)
	if errs := h.V.Validate(in, rules...); !errs.Empty() {
		response.ValidationErrors(c, 400, response.MsgFailedUpdate, errs)
		return
	}

	if f := in.File("image"); f != nil {
		if g.Image != "" {
			h.Storage.DeleteByURL(storage.GalleryDir, g.Image)
		}
		url, err := h.Storage.SaveURL(f, storage.GalleryDir)
		if err != nil {
			h.serverError(c, err)
			return
		}
		g.Image = url
	} else if in.String("remove_existing_image") == "true" {
		if g.Image != "" {
			h.Storage.DeleteByURL(storage.GalleryDir, g.Image)
		}
		g.Image = ""
	}

	g.Caption = in.Str("caption")
	if in.Has("latitude") {
		g.Latitude = in.Str("latitude")
	}
	if in.Has("longitude") {
		g.Longitude = in.Str("longitude")
	}
	g.Slug = gallerySlug(g.Caption)

	if err := h.DB.Select("*").Omit("created_at").Save(&g).Error; err != nil {
		h.serverError(c, err)
		return
	}
	h.attachUser(&g)
	response.Success(c, 200, response.MsgSuccessUpdate, g)
}

// DeleteGallery — DELETE /content/gallery/:id (admin, officer, member)
func (h *Handler) DeleteGallery(c *gin.Context) {
	var g models.Gallery
	if err := h.DB.First(&g, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	if err := h.DB.Delete(&g).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, response.MsgSuccessDelete)
}

// ShowImage — GET /images/*path : serves a file from the public disk.
func (h *Handler) ShowImage(c *gin.Context) {
	p, ok := h.Storage.Resolve(c.Param("path"))
	if !ok {
		response.Error(c, 404, response.MsgNoItem, nil, 0)
		return
	}
	c.File(p)
}

// attachUser loads the owner without re-reading the gallery row, so the
// response echoes the submitted values (e.g. latitude "-6.9") like Laravel.
func (h *Handler) attachUser(g *models.Gallery) {
	if g.UserID != nil {
		var u models.User
		if h.DB.First(&u, "id = ?", *g.UserID).Error == nil {
			g.User = &u
		}
	}
}

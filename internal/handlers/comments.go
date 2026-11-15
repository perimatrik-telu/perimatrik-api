package handlers

import (
	"perimatrik-api-go/internal/auth"
	"perimatrik-api-go/internal/models"
	"perimatrik-api-go/internal/response"
	"perimatrik-api-go/internal/validation"

	"github.com/gin-gonic/gin"
)

// ListComments — GET /content/comments?article_id=|news_id=
// Returns top-level comments (newest first) with nested replies.
func (h *Handler) ListComments(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("article_id", "nullable|uuid|exists:articles,id"),
		validation.R("news_id", "nullable|uuid|exists:news,id"),
	)
	if !errs.Empty() {
		response.Message(c, 422, false, "Invalid ID provided.")
		return
	}

	q := h.DB.Preload("User").Where("parent_id IS NULL").Order("created_at desc")
	switch {
	case in.Filled("article_id"):
		q = q.Where("article_id = ?", in.String("article_id"))
	case in.Filled("news_id"):
		q = q.Where("news_id = ?", in.String("news_id"))
	default:
		response.Message(c, 400, false, "Context ID is required.")
		return
	}

	roots := []*models.Comment{}
	if err := q.Find(&roots).Error; err != nil {
		h.serverError(c, err)
		return
	}
	if err := h.loadReplies(roots); err != nil {
		h.serverError(c, err)
		return
	}
	response.Data(c, 200, roots)
}

// loadReplies fills Replies recursively, one query per nesting level.
func (h *Handler) loadReplies(level []*models.Comment) error {
	for len(level) > 0 {
		byID := make(map[string]*models.Comment, len(level))
		ids := make([]string, 0, len(level))
		for _, c := range level {
			c.Replies = []*models.Comment{}
			byID[c.ID] = c
			ids = append(ids, c.ID)
		}
		var children []*models.Comment
		if err := h.DB.Preload("User").Where("parent_id IN ?", ids).Order("created_at asc").Find(&children).Error; err != nil {
			return err
		}
		for _, ch := range children {
			if p := byID[*ch.ParentID]; p != nil {
				p.Replies = append(p.Replies, ch)
			}
		}
		level = children
	}
	return nil
}

// CreateComment — POST /content/comments (auth)
func (h *Handler) CreateComment(c *gin.Context) {
	in := h.input(c)
	if in == nil {
		return
	}
	errs := h.V.Validate(in,
		validation.R("body", "required|string"),
		validation.R("article_id", "nullable|uuid|exists:articles,id"),
		validation.R("news_id", "nullable|uuid|exists:news,id"),
		validation.R("parent_id", "nullable|uuid|exists:comments,id"),
	)
	if !errs.Empty() {
		response.ValidationErrors(c, 422, "Validation failed", errs)
		return
	}

	cm := models.Comment{
		Body:      in.String("body"),
		UserID:    auth.MustUser(c).ID,
		ArticleID: in.Str("article_id"),
		NewsID:    in.Str("news_id"),
		ParentID:  in.Str("parent_id"),
	}
	if err := h.DB.Create(&cm).Error; err != nil {
		h.serverError(c, err)
		return
	}
	cm.User = auth.MustUser(c)
	cm.Replies = []*models.Comment{}
	response.Data(c, 201, cm)
}

// DeleteComment — DELETE /content/comments/:id (owner or admin)
func (h *Handler) DeleteComment(c *gin.Context) {
	var cm models.Comment
	if err := h.DB.First(&cm, "id = ?", c.Param("id")).Error; err != nil {
		if notFound(err) {
			response.Message(c, 404, false, response.MsgNoItem)
		} else {
			h.serverError(c, err)
		}
		return
	}
	me := auth.MustUser(c)
	if !isAdmin(me) && me.ID != cm.UserID {
		response.Message(c, 403, false, "Unauthorized to delete this comment.")
		return
	}
	if err := h.DB.Delete(&cm).Error; err != nil {
		h.serverError(c, err)
		return
	}
	response.Message(c, 200, true, "Comment deleted successfully.")
}

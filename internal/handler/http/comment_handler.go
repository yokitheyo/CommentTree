package http

import (
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/domain"
	"github.com/yokitheyo/CommentTree/internal/dto"
)

type CommentHandler struct {
	service domain.CommentService
}

func NewCommentHandler(service domain.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) RegisterRoutes(engine *ginext.Engine) {
	group := engine.Group("/comments")
	group.POST("", h.CreateComment)
	group.GET("", h.GetComments)
	group.DELETE("/:id", h.DeleteComment)
	group.GET("/search", h.SearchComments)
}

// CreateComment POST /comments
func (h *CommentHandler) CreateComment(c *ginext.Context) {
	var req dto.CreateCommentRequest
	if err := c.BindJSON(&req); err != nil {
		zlog.Logger.Warn().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid request"})
		return
	}

	log := zlog.Logger.Debug().Str("author", req.Author)
	if req.ParentID != nil {
		log = log.Int64("parent_id", *req.ParentID)
	}
	log.Msg("CreateComment called")

	comment, err := h.service.CreateComment(c, req.ParentID, req.Author, req.Content)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("CreateComment failed")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to create comment"})
		return
	}

	zlog.Logger.Debug().Int64("comment_id", comment.ID).Msg("comment created successfully")
	c.JSON(http.StatusCreated, MapToCommentResponse(comment))

}

// GetComments GET /comments?parent={id}&limit=&offset=&sort=
func (h *CommentHandler) GetComments(c *ginext.Context) {
	var parentID *int64
	if parentStr := c.Query("parent"); parentStr != "" {
		id, err := strconv.ParseInt(parentStr, 10, 64)
		if err != nil {
			zlog.Logger.Warn().Err(err).Str("parent", parentStr).Msg("invalid parent id")
			c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid parent id"})
			return
		}
		parentID = &id
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		} else {
			zlog.Logger.Warn().Err(err).Str("limit", l).Msg("invalid limit parameter, using default")
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = val
		} else {
			zlog.Logger.Warn().Err(err).Str("offset", o).Msg("invalid offset parameter, using default")
		}
	}
	sort := c.Query("sort")

	log := zlog.Logger.Debug().Int("limit", limit).Int("offset", offset).Str("sort", sort)
	if parentID != nil {
		log = log.Int64("parent_id", *parentID)
	}
	log.Msg("GetComments called with parameters")

	comments, err := h.service.GetThread(c, parentID, limit, offset, sort)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("GetThread failed")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get comments"})
		return
	}

	c.JSON(http.StatusOK, MapToCommentResponses(comments))
}

// DeleteComment DELETE /comments/:id
func (h *CommentHandler) DeleteComment(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		zlog.Logger.Warn().Err(err).Str("id", idStr).Msg("invalid id parameter")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid id"})
		return
	}

	zlog.Logger.Debug().Int64("comment_id", id).Msg("DeleteThread called")

	if err := h.service.DeleteThread(c, id); err != nil {
		zlog.Logger.Error().Err(err).Int64("comment_id", id).Msg("DeleteThread failed")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to delete comment"})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchComments GET /comments/search?query=&limit=&offset=
func (h *CommentHandler) SearchComments(c *ginext.Context) {
	query := c.Query("query")
	if query == "" {
		zlog.Logger.Warn().Msg("search query is empty")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "query cannot be empty"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		} else {
			zlog.Logger.Warn().Err(err).Str("limit", l).Msg("invalid limit parameter in search, using default")
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = val
		} else {
			zlog.Logger.Warn().Err(err).Str("offset", o).Msg("invalid offset parameter in search, using default")
		}
	}

	zlog.Logger.Debug().Str("query", query).Int("limit", limit).Int("offset", offset).Msg("SearchComment called")

	comments, err := h.service.SearchComment(c, query, limit, offset)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("query", query).Msg("SearchComment failed")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

package http

import (
	"github.com/yokitheyo/CommentTree/internal/domain"
	"github.com/yokitheyo/CommentTree/internal/dto"
)

func MapToCommentResponse(c *domain.Comment) *dto.CommentResponse {
	if c == nil {
		return nil
	}

	children := make([]*dto.CommentResponse, 0, len(c.Children))
	for _, ch := range c.Children {
		children = append(children, MapToCommentResponse(ch))
	}

	return &dto.CommentResponse{
		ID:        c.ID,
		ParentID:  c.ParentID,
		Content:   c.Content,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Deleted:   c.Deleted,
		Children:  children,
	}
}

func MapToCommentResponses(list []*domain.Comment) []*dto.CommentResponse {
	out := make([]*dto.CommentResponse, 0, len(list))
	for _, c := range list {
		out = append(out, MapToCommentResponse(c))
	}
	return out
}

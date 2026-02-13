package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/infrastructure/search"

	"github.com/yokitheyo/CommentTree/internal/domain"
)

type CommentUsecase struct {
	repo   domain.CommentRepository
	search search.FullTextSearcher
}

func NewCommentUsecase(repo domain.CommentRepository, search search.FullTextSearcher) *CommentUsecase {
	return &CommentUsecase{
		repo:   repo,
		search: search,
	}
}

func (u *CommentUsecase) CreateComment(ctx context.Context, parentID *int64, author, content string) (*domain.Comment, error) {
	if author == "" {
		return nil, errors.New("author required")
	}
	if content == "" {
		return nil, errors.New("content required")
	}

	c := &domain.Comment{
		ParentID: parentID,
		Author:   author,
		Content:  content,
	}

	if err := u.repo.Save(ctx, c); err != nil {
		zlog.Logger.Error().Err(err).Msg("usecase: Save comment failed")
		return nil, fmt.Errorf("save comment: %w", err)
	}

	zlog.Logger.Info().Msgf("comment created id=%d parent=%v", c.ID, c.ParentID)
	return c, nil
}

func (u *CommentUsecase) GetThread(ctx context.Context, parentID *int64, limit, offset int, sort string) ([]*domain.Comment, error) {
	comments, err := u.repo.FindChildren(ctx, parentID, limit, offset, sort)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("usecase: FindChildren failed")
		return nil, fmt.Errorf("find children for parent_id=%v: %w", parentID, err)
	}

	zlog.Logger.Info().Msgf("GetThread found %d comments for parent_id=%v", len(comments), parentID)

	for _, comment := range comments {
		if err := u.loadAllChildren(ctx, comment); err != nil {
			zlog.Logger.Error().Err(err).Msgf("failed to load children for comment %d", comment.ID)
		}
	}

	return comments, nil
}

func (u *CommentUsecase) loadAllChildren(ctx context.Context, comment *domain.Comment) error {
	children, err := u.repo.FindChildren(ctx, &comment.ID, 1000, 0, "asc")
	if err != nil {
		return fmt.Errorf("load children for comment %d: %w", comment.ID, err)
	}

	comment.Children = children
	zlog.Logger.Debug().Msgf("loaded %d children for comment %d", len(children), comment.ID)

	for _, child := range children {
		if err := u.loadAllChildren(ctx, child); err != nil {
			zlog.Logger.Error().Err(err).Msgf("failed to load children for comment %d", child.ID)
		}
	}

	return nil
}

func (u *CommentUsecase) DeleteThread(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	if err := u.repo.Delete(ctx, id); err != nil {
		zlog.Logger.Error().Err(err).Msgf("usecase: Delete failed id=%d", id)
		return fmt.Errorf("delete comment id=%d: %w", id, err)
	}
	zlog.Logger.Info().Msgf("comment deleted id=%d", id)
	return nil
}

func (u *CommentUsecase) SearchComment(ctx context.Context, q string, limit, offset int) ([]*domain.Comment, error) {
	if q == "" {
		return nil, errors.New("empty query")
	}
	comments, err := u.search.SearchComments(ctx, q, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search comments: %w", err)
	}
	return comments, nil
}

package search

import (
	"context"

	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/domain"
)

type FullTextSearcher interface {
	SearchComments(ctx context.Context, query string, limit, offset int) ([]*domain.Comment, error)
}

type PostgresFullText struct {
	repo domain.CommentRepository
}

func NewPostgresFullText(repo domain.CommentRepository) *PostgresFullText {
	return &PostgresFullText{repo: repo}
}

func (f *PostgresFullText) SearchComments(ctx context.Context, query string, limit, offset int) ([]*domain.Comment, error) {
	zlog.Logger.Debug().Str("query", query).Int("limit", limit).Int("offset", offset).Msg("search: SearchComments starting")

	comments, err := f.repo.Search(ctx, query, limit, offset)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("query", query).Msg("search: SearchComments failed")
		return nil, err
	}

	zlog.Logger.Info().Str("query", query).Int("results", len(comments)).Msg("search: SearchComments completed")
	return comments, nil
}

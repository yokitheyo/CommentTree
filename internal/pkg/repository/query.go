package repository

import (
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/domain"
)

func QueryComments(ctx context.Context, db *dbpg.DB, strategy retry.Strategy, query string, args ...interface{}) ([]*domain.Comment, error) {
	rows, err := db.QueryWithRetry(ctx, strategy, query, args...)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("query failed")
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		c, err := ScanComment(rows)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("scan failed")
			return nil, err
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		zlog.Logger.Error().Err(err).Msg("rows iteration failed")
		return nil, err
	}

	return comments, nil
}

func OrderByCreated(sort string) string {
	if sort == "desc" {
		return "created_at DESC"
	}
	return "created_at ASC"
}

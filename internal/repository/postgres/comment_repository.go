package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/domain"
	"github.com/yokitheyo/CommentTree/internal/pkg/repository"
)

type commentRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewCommentRepository(db *dbpg.DB, strategy retry.Strategy) domain.CommentRepository {
	return &commentRepository{db: db, strategy: strategy}
}

func (r *commentRepository) Save(ctx context.Context, c *domain.Comment) error {
	query := `
    INSERT INTO comments (parent_id, author, content, deleted)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, updated_at
`
	err := r.db.Master.QueryRowContext(ctx, query,
		c.ParentID,
		c.Author,
		c.Content,
		c.Deleted,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("author", c.Author).Msg("repository: Save comment failed")
		return fmt.Errorf("save comment: %w", err)
	}

	log := zlog.Logger.Debug().Int64("comment_id", c.ID)
	if c.ParentID != nil {
		log = log.Int64("parent_id", *c.ParentID)
	}
	log.Msg("comment saved to database")
	return nil
}

func (r *commentRepository) FindByID(ctx context.Context, id int64) (*domain.Comment, error) {
	query := `
		SELECT id, parent_id, author, content, created_at, updated_at, deleted
		FROM comments
		WHERE id = $1
	`

	row := r.db.Master.QueryRowContext(ctx, query, id)
	c, err := repository.ScanComment(row)
	if err != nil {
		if err == sql.ErrNoRows {
			zlog.Logger.Debug().Int64("comment_id", id).Msg("comment not found")
			return nil, fmt.Errorf("comment id=%d: %w", id, domain.ErrCommentNotFound)
		}
		zlog.Logger.Error().Err(err).Int64("comment_id", id).Msg("repository: FindByID failed")
		return nil, fmt.Errorf("find comment by id=%d: %w", id, err)
	}

	zlog.Logger.Debug().Int64("comment_id", id).Msg("comment found by id")
	return c, nil
}

func (r *commentRepository) FindChildren(ctx context.Context, parentID *int64, limit, offset int, sort string) ([]*domain.Comment, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = fmt.Sprintf(`
			SELECT id, parent_id, author, content, created_at, updated_at, deleted
			FROM comments
			WHERE parent_id IS NULL AND deleted = false
			ORDER BY %s
			LIMIT $1 OFFSET $2
		`, repository.OrderByCreated(sort))
		args = []interface{}{limit, offset}
	} else {
		query = fmt.Sprintf(`
			SELECT id, parent_id, author, content, created_at, updated_at, deleted
			FROM comments
			WHERE parent_id = $1 AND deleted = false
			ORDER BY %s
			LIMIT $2 OFFSET $3
		`, repository.OrderByCreated(sort))
		args = []interface{}{*parentID, limit, offset}
	}

	log := zlog.Logger.Debug().Int("limit", limit).Int("offset", offset)
	if parentID != nil {
		log = log.Int64("parent_id", *parentID)
	}
	log.Msg("repository: FindChildren query starting")

	comments, err := repository.QueryComments(ctx, r.db, r.strategy, query, args...)
	if err != nil {
		zlog.Logger.Error().Err(err).Interface("parent_id", parentID).Msg("repository: FindChildren failed")
		return nil, fmt.Errorf("find children parent_id=%v: %w", parentID, err)
	}

	log = zlog.Logger.Debug().Int("count", len(comments))
	if parentID != nil {
		log = log.Int64("parent_id", *parentID)
	}
	log.Msg("repository: FindChildren completed")
	return comments, nil
}

func (r *commentRepository) Delete(ctx context.Context, id int64) error {
	zlog.Logger.Debug().Int64("comment_id", id).Msg("repository: Delete starting")

	res, err := r.db.ExecWithRetry(ctx, r.strategy, `
		UPDATE comments
		SET deleted = true, updated_at = $2
		WHERE id = $1
	`, id, time.Now())

	if err != nil {
		zlog.Logger.Error().Err(err).Int64("comment_id", id).Msg("repository: Delete failed")
		return fmt.Errorf("delete comment id=%d: %w", id, err)
	}

	zlog.Logger.Debug().Int64("comment_id", id).Interface("result", res).Msg("comment marked as deleted")
	return nil
}

func (r *commentRepository) Search(ctx context.Context, q string, limit, offset int) ([]*domain.Comment, error) {
	query := `
		SELECT id, parent_id, author, content, created_at, updated_at, deleted
		FROM comments
		WHERE (content ILIKE '%' || $1 || '%' OR author ILIKE '%' || $1 || '%')
		AND deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	zlog.Logger.Debug().Str("search_query", q).Int("limit", limit).Int("offset", offset).Msg("repository: Search query starting")

	comments, err := repository.QueryComments(ctx, r.db, r.strategy, query, q, limit, offset)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("search_query", q).Msg("repository: Search failed")
		return nil, fmt.Errorf("search comments query=%q: %w", q, err)
	}

	zlog.Logger.Debug().Str("search_query", q).Int("count", len(comments)).Msg("repository: Search completed")
	return comments, nil
}

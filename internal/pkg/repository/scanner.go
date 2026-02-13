package repository

import (
	"database/sql"
	"fmt"

	"github.com/yokitheyo/CommentTree/internal/domain"
)

// RowScanner интерфейс для сканирования строк (покрывает *sql.Row и *sql.Rows)
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// ScanComment сканирует комментарий из строки БД
func ScanComment(row RowScanner) (*domain.Comment, error) {
	c := &domain.Comment{}
	var parent sql.NullInt64
	var updated sql.NullTime

	if err := row.Scan(&c.ID, &parent, &c.Author, &c.Content, &c.CreatedAt, &updated, &c.Deleted); err != nil {
		return nil, fmt.Errorf("scan comment: %w", err)
	}

	if parent.Valid {
		c.ParentID = &parent.Int64
	}
	if updated.Valid {
		c.UpdatedAt = &updated.Time
	}

	return c, nil
}

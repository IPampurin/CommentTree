package db

import (
	"context"
	"fmt"
)

const (
	// commentsSchema создаёт таблицу comments с поддержкой древовидной структуры (parent_id NULL для корневых)
	commentsSchema = `CREATE TABLE IF NOT EXISTS comments (
			              id SERIAL PRIMARY KEY,
			       parent_id INT NULL,
			            text TEXT NOT NULL,
			      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE
		);

		-- Индекс для быстрого получения дочерних комментариев по parent_id
		CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);

		-- Индекс для сортировки по дате создания
		CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);

		-- Индекс для полнотекстового поиска (русский + английский)
		CREATE INDEX IF NOT EXISTS idx_comments_text_gin ON comments USING GIN (to_tsvector('russian', text) || to_tsvector('english', text));`
)

// Migration создаёт таблицу comments и необходимые индексы, если они ещё не существуют
func (d *DataBase) Migration(ctx context.Context) error {

	_, err := d.Pool.Exec(ctx, commentsSchema)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы comments и индексов: %w", err)
	}

	return nil
}

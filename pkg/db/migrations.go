package db

import (
	"context"
	"fmt"
)

// Migration создаёт таблицу comments и необходимые индексы, если они ещё не существуют
func (d *DataBase) Migration(ctx context.Context) error {

	commands := []string{
		`CREATE TABLE IF NOT EXISTS comments (
	         id SERIAL PRIMARY KEY,
      parent_id INT NULL,
           text TEXT NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_text_gin_ru ON comments USING GIN (to_tsvector('russian', text));`,
		`CREATE INDEX IF NOT EXISTS idx_comments_text_gin_en ON comments USING GIN (to_tsvector('english', text));`,
	}

	for _, cmd := range commands {
		_, err := d.Pool.Exec(ctx, cmd)
		if err != nil {
			return fmt.Errorf("ошибка создания таблицы comments и индексов: %w", err)
		}
	}

	return nil
}

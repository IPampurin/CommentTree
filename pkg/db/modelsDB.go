package db

import (
	"time"
)

// Comment представляет запись из таблицы comments
type Comment struct {
	ID        int       `json:"id"`         // идентификатор комментария (автоинкремент)
	ParentID  *int      `json:"parent_id"`  // ID родительского комментария, NULL для корневых
	Text      string    `json:"text"`       // текст комментария
	CreatedAt time.Time `json:"created_at"` // дата и время создания комментария
}

// CommentNode используется для построения дерева комментариев (с детьми)
type CommentNode struct {
	Comment  *Comment       `json:"comment"`
	Children []*CommentNode `json:"children"`
}

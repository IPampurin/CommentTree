package service

import "time"

// Comment представляет комментарий для бизнес-логики (аналог db.Comment)
type Comment struct {
	ID        int       `json:"id"`         // идентификатор комментария
	ParentID  *int      `json:"parent_id"`  // ID родителя, nil для корневых
	Text      string    `json:"text"`       // текст комментария
	CreatedAt time.Time `json:"created_at"` // дата создания
}

// CommentNode используется для построения дерева комментариев
type CommentNode struct {
	Comment  *Comment       `json:"comment"`  // сам комментарий
	Children []*CommentNode `json:"children"` // дочерние комментарии
}

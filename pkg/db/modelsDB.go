package db

import (
	"time"
)

// Comment представляет запись о комментарии
type Comment struct {
	ID            int        // идентификатор комментария (автоинкремент)
	ParentID      int        // ID родительского комментария, 0 для корневых
	Text          string     // текст комментария
	CreatedAt     time.Time  // дата и время создания комментария
	ChildComments []*Comment // дочерние комментарии
}

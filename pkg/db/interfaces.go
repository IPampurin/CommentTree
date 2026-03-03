package db

import "context"

// методы для работы с базой комментариев
type CommentsMethods interface {

	// добавляет запись в БД
	CreateComment(ctx context.Context, parentID *int, text string) (int, error)

	// возвращает запись по ID
	GetCommentByID(ctx context.Context, id int) (*Comment, error)

	// возвращает список прямых потомков родителя
	GetCommentsByParent(ctx context.Context, parentID *int, page, limit int, sortBy string) ([]*Comment, int, error)

	// возвращает всё поддерево для указанного корневого комментария
	GetCommentTree(ctx context.Context, rootID int) (*CommentNode, error)

	// удаляет запись из БД со всеми зависимыми
	DeleteComment(ctx context.Context, id int) error

	// возвращает список записей содержащих поисковый запрос
	SearchComments(ctx context.Context, query string, page, limit int) ([]*Comment, int, error)
}

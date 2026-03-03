package service

import "context"

// CommentService определяет методы бизнес-логики для работы с комментариями
type CommentService interface {

	// CreateComment создаёт новый комментарий (parentID может быть nil (корневой комментарий))
	CreateComment(ctx context.Context, parentID *int, text string) (int, error)

	// GetComment возвращает комментарий по ID (без дерева)
	GetComment(ctx context.Context, id int) (*Comment, error)

	// GetCommentsByParent возвращает плоский список комментариев с пагинацией и сортировкой (parentID = nil означает корневые комментарии)
	GetCommentsByParent(ctx context.Context, parentID *int, page, limit int, sortBy string) ([]*Comment, int, error)

	// GetCommentTree возвращает дерево комментариев для указанного корневого ID (если комментарий не найден, возвращает nil, nil)
	GetCommentTree(ctx context.Context, rootID int) (*CommentNode, error)

	// DeleteComment удаляет комментарий и всех его потомков
	DeleteComment(ctx context.Context, id int) error

	// SearchComments выполняет полнотекстовый поиск и возвращает плоский список с пагинацией
	SearchComments(ctx context.Context, query string, page, limit int) ([]*Comment, int, error)
}

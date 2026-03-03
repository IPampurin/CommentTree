package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/IPampurin/CommentTree/pkg/db"
)

// CreateComment создаёт комментарий через БД
func (s *Service) CreateComment(ctx context.Context, parentID *int, text string) (int, error) {

	if text == "" {
		return 0, errors.New("текст комментария не может быть пустым")
	}

	id, err := s.storage.CreateComment(ctx, parentID, text)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания комментария: %w", err)
	}

	return id, nil
}

// GetComment возвращает комментарий по ID
func (s *Service) GetComment(ctx context.Context, id int) (*Comment, error) {

	comment, err := s.storage.GetCommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения комментария: %w", err)
	}
	if comment == nil {
		return nil, nil
	}

	// преобразуем db.Comment в service.Comment
	return &Comment{
		ID:        comment.ID,
		ParentID:  comment.ParentID,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt,
	}, nil
}

// GetCommentsByParent возвращает плоский список с пагинацией
func (s *Service) GetCommentsByParent(ctx context.Context, parentID *int, page, limit int, sortBy string) ([]*Comment, int, error) {

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10 // значение по умолчанию
	}

	dbComments, total, err := s.storage.GetCommentsByParent(ctx, parentID, page, limit, sortBy)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения комментариев: %w", err)
	}

	// преобразуем
	comments := make([]*Comment, len(dbComments))
	for i, c := range dbComments {
		comments[i] = &Comment{
			ID:        c.ID,
			ParentID:  c.ParentID,
			Text:      c.Text,
			CreatedAt: c.CreatedAt,
		}
	}

	return comments, total, nil
}

// GetCommentTree получает дерево комментариев
func (s *Service) GetCommentTree(ctx context.Context, rootID int) (*CommentNode, error) {

	node, err := s.storage.GetCommentTree(ctx, rootID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения дерева комментариев: %w", err)
	}

	if node == nil {
		return nil, nil
	}

	// рекурсивное преобразование db.CommentNode в service.CommentNode
	return s.convertNode(node), nil
}

// вспомогательная рекурсивная функция преобразования дерева
func (s *Service) convertNode(dbNode *db.CommentNode) *CommentNode {

	if dbNode == nil {
		return nil
	}

	svcNode := &CommentNode{
		Comment: &Comment{
			ID:        dbNode.Comment.ID,
			ParentID:  dbNode.Comment.ParentID,
			Text:      dbNode.Comment.Text,
			CreatedAt: dbNode.Comment.CreatedAt,
		},
		Children: make([]*CommentNode, len(dbNode.Children)),
	}

	for i, child := range dbNode.Children {
		svcNode.Children[i] = s.convertNode(child)
	}

	return svcNode
}

// DeleteComment удаляет комментарий
func (s *Service) DeleteComment(ctx context.Context, id int) error {

	err := s.storage.DeleteComment(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления комментария: %w", err)
	}

	return nil
}

// SearchComments выполняет поиск query
func (s *Service) SearchComments(ctx context.Context, query string, page, limit int) ([]*Comment, int, error) {

	if query == "" {
		return nil, 0, errors.New("пустой поисковый запрос")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	dbComments, total, err := s.storage.SearchComments(ctx, query, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка поиска комментариев: %w", err)
	}

	comments := make([]*Comment, len(dbComments))

	for i, c := range dbComments {
		comments[i] = &Comment{
			ID:        c.ID,
			ParentID:  c.ParentID,
			Text:      c.Text,
			CreatedAt: c.CreatedAt,
		}
	}

	return comments, total, nil
}

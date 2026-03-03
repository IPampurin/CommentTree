package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// CreateComment добавляет новый комментарий в БД (parentID может быть nil (корневой комментарий))
func (d *DataBase) CreateComment(ctx context.Context, parentID *int, text string) (int, error) {

	query := `   INSERT INTO comments (parent_id, text)
	             VALUES ($1, $2)
		      RETURNING id`

	var id int

	err := d.Pool.QueryRow(ctx, query, parentID, text).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка CreateComment при добавлении комментария в БД: %w", err)
	}

	return id, nil
}

// GetCommentByID возвращает комментарий по его ID
func (d *DataBase) GetCommentByID(ctx context.Context, id int) (*Comment, error) {

	query := `SELECT id, parent_id, text, created_at
	            FROM comments
			   WHERE id = $1`

	row := d.Pool.QueryRow(ctx, query, id)

	var c Comment
	err := row.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // комментарий не найден
		}

		return nil, fmt.Errorf("ошибка GetCommentByID при получении комментария в БД: %w", err)
	}

	return &c, nil
}

// GetCommentsByParent возвращает список прямых потомков для заданного родителя с пагинацией и сортировкой (parentID может быть nil (корневые комментарии))
// Параметры:
//   - page: номер страницы (начиная с 1)
//   - limit: количество записей на странице
//   - sortBy: поле для сортировки (например, "created_at DESC" или "created_at ASC")
//
// Возвращает:
//   - список комментариев
//   - общее количество записей (для пагинации)
//   - ошибка
func (d *DataBase) GetCommentsByParent(ctx context.Context, parentID *int, page, limit int, sortBy string) ([]*Comment, int, error) {

	// безопасная сортировка: разрешим только известные поля
	allowedSort := map[string]bool{
		"created_at ASC":  true,
		"created_at DESC": true,
	}
	if !allowedSort[sortBy] {
		sortBy = "created_at DESC" // значение по умолчанию
	}

	offset := (page - 1) * limit

	// получаем общее количество
	countQuery := `SELECT COUNT(*)
	                 FROM comments
				    WHERE parent_id IS NOT DISTINCT FROM $1` // корректно обрабатывает NULL

	var total int

	err := d.Pool.QueryRow(ctx, countQuery, parentID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка GetCommentsByParent count при получении количества записей: %w", err)
	}

	// выбираем данные с пагинацией
	pagQuery := fmt.Sprintf(`SELECT id, parent_id, text, created_at
		                       FROM comments
		                      WHERE parent_id IS NOT DISTINCT FROM $1
		                      ORDER BY %s
		                      LIMIT $2 OFFSET $3`, sortBy)

	rows, err := d.Pool.Query(ctx, pagQuery, parentID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка GetCommentsByParent query при выборке с пагинацией: %w", err)
	}
	defer rows.Close()

	var comments []*Comment

	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("ошибка GetCommentsByParent scan при применении метода Scan: %w", err)
		}

		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка GetCommentsByParent rows при итерации по выборке: %w", err)
	}

	return comments, total, nil
}

// GetCommentTree рекурсивно загружает всё поддерево для указанного корневого комментария,
// возвращает узел с детьми, если комментарий не найден, возвращает (nil, nil)
func (d *DataBase) GetCommentTree(ctx context.Context, rootID int) (*CommentNode, error) {

	// получаем корневой комментарий
	rootComment, err := d.GetCommentByID(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if rootComment == nil {
		return nil, nil
	}

	// рекурсивно загружаем детей
	children, err := d.getChildNodes(ctx, rootID)
	if err != nil {
		return nil, err
	}

	return &CommentNode{
		Comment:  rootComment,
		Children: children,
	}, nil
}

// getChildNodes – вспомогательная рекурсивная функция для загрузки всех потомков узла
func (d *DataBase) getChildNodes(ctx context.Context, parentID int) ([]*CommentNode, error) {

	query := `SELECT id, parent_id, text, created_at
		        FROM comments
		       WHERE parent_id = $1
		       ORDER BY created_at ASC` // внутри ветки сортируем по возрастанию (старые сверху)

	rows, err := d.Pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("ошибка getChildNodes query при выборке: %w", err)
	}
	defer rows.Close()

	var nodes []*CommentNode
	for rows.Next() {

		var c Comment
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка getChildNodes scan при применении метода Scan: %w", err)
		}
		// рекурсивно загружаем внуков
		grandChildren, err := d.getChildNodes(ctx, c.ID)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, &CommentNode{
			Comment:  &c,
			Children: grandChildren,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка getChildNodes rows при итерации по выборке: %w", err)
	}

	return nodes, nil
}

// DeleteComment удаляет комментарий и всех его потомков
func (d *DataBase) DeleteComment(ctx context.Context, id int) error {

	query := `DELETE
	            FROM comments
			   WHERE id = $1`

	commandTag, err := d.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка DeleteComment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("comment not found") // применим errors.New() для возможности дальнейшей детальной обработки
	}

	return nil
}

// SearchComments выполняет полнотекстовый поиск с использованием plainto_tsquery,
// возвращает плоский список найденных комментариев (без иерархии) с пагинацией
func (d *DataBase) SearchComments(ctx context.Context, query string, page, limit int) ([]*Comment, int, error) {

	offset := (page - 1) * limit

	// общее количество
	countQuery := `SELECT COUNT(*)
                     FROM comments
                    WHERE to_tsvector('russian', text) @@ plainto_tsquery('russian', $1)
                          OR to_tsvector('english', text) @@ plainto_tsquery('english', $2)`

	var total int

	err := d.Pool.QueryRow(ctx, countQuery, query, query).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка SearchComments count при выборке: %w", err)
	}

	// результаты с пагинацией
	selectQuery := `SELECT id, parent_id, text, created_at
                      FROM comments
                     WHERE to_tsvector('russian', text) @@ plainto_tsquery('russian', $1)
                           OR to_tsvector('english', text) @@ plainto_tsquery('english', $2)
                     ORDER BY created_at DESC
                     LIMIT $3 OFFSET $4`

	rows, err := d.Pool.Query(ctx, selectQuery, query, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка SearchComments query при получении результата с пагинацией: %w", err)
	}
	defer rows.Close()

	var comments []*Comment

	for rows.Next() {

		var c Comment
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("ошибка SearchComments scan при применении метода Scan: %w", err)
		}

		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка SearchComments rows при итерации по выборке: %w", err)
	}

	return comments, total, nil
}

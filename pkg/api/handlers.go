package api

import (
	"net/http"
	"strconv"

	"github.com/IPampurin/CommentTree/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/logger"
)

// CreateComment обработчик POST /comments
func CreateComment(svc service.CommentService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req createCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Warn("некорректный JSON при создании комментария", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос: " + err.Error()})
			return
		}

		id, err := svc.CreateComment(c.Request.Context(), req.ParentID, req.Text)
		if err != nil {
			log.Error("ошибка создания комментария", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать комментарий"})
			return
		}

		// возвращаем ID созданного комментария
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

// GetComments обработчик GET /comments
// Поддерживает query-параметры:
//   - parent (int, опционально) – ID родителя, если не указан – корневые
//   - page (int, по умолчанию 1)
//   - limit (int, по умолчанию 10)
//   - sort (string, по умолчанию "created_at DESC") – допустимые значения: created_at ASC, created_at DESC
func GetComments(svc service.CommentService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		// парсим parent
		var parentID *int
		if parentStr := c.Query("parent"); parentStr != "" {
			if pid, err := strconv.Atoi(parentStr); err == nil {
				parentID = &pid
			} else {
				log.Warn("некорректный parent ID", "value", parentStr)
				c.JSON(http.StatusBadRequest, gin.H{"error": "parent должен быть целым числом"})
				return
			}
		}

		// парсим page и limit
		page := 1
		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		limit := 10
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		// сортируем
		sortBy := c.DefaultQuery("sort", "created_at DESC")
		allowedSorts := map[string]bool{"created_at ASC": true, "created_at DESC": true}
		if !allowedSorts[sortBy] {
			sortBy = "created_at DESC"
		}

		// если указан конкретный parentID и это не корневой запрос - отдаём дерево
		// если parent указан - возвращаем дерево
		if parentID != nil {
			tree, err := svc.GetCommentTree(c.Request.Context(), *parentID)
			if err != nil {
				log.Error("ошибка получения дерева комментариев", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить комментарии"})
				return
			}
			if tree == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "комментарий не найден"})
				return
			}

			// преобразуем дерево в response
			resp := convertNodeToResponse(tree)
			c.JSON(http.StatusOK, resp)
			return
		}

		// иначе возвращаем корневые комментарии в виде деревьев с пагинацией
		// получаем плоский список корневых комментариев
		rootComments, total, err := svc.GetCommentsByParent(c.Request.Context(), nil, page, limit, sortBy)
		if err != nil {
			log.Error("ошибка получения корневых комментариев", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить комментарии"})
			return
		}

		// для каждого корневого комментария загружаем полное дерево
		trees := make([]*treeResponse, 0, len(rootComments))
		for _, comm := range rootComments {
			tree, err := svc.GetCommentTree(c.Request.Context(), comm.ID)
			if err != nil {
				log.Error("ошибка получения дерева для комментария", "id", comm.ID, "error", err)
				// пропускаем этот комментарий, чтобы не ломать весь список
				continue
			}
			if tree != nil {
				trees = append(trees, convertNodeToResponse(tree))
			}
		}

		// отвечаем объектом, содержащим массив деревьев и метаданные пагинации
		c.JSON(http.StatusOK, gin.H{
			"comments": trees,
			"total":    total,
			"page":     page,
			"limit":    limit,
		})
	}
}

// DeleteComment обработчик DELETE /comments/:id
func DeleteComment(svc service.CommentService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Warn("некорректный ID комментария", "value", idStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
			return
		}

		err = svc.DeleteComment(c.Request.Context(), id)
		if err != nil {
			log.Error("ошибка удаления комментария", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить комментарий"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// SearchComments обработчик GET /comments/search
// Параметры:
//   - q (string, обязательный) – поисковый запрос
//   - page (int, по умолчанию 1)
//   - limit (int, по умолчанию 10)
func SearchComments(svc service.CommentService, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "параметр q (поисковый запрос) обязателен"})
			return
		}

		page := 1
		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		limit := 10
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		comments, total, err := svc.SearchComments(c.Request.Context(), query, page, limit)
		if err != nil {
			log.Error("ошибка поиска комментариев", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось выполнить поиск"})
			return
		}

		respComments := make([]*commentResponse, len(comments))
		for i, comm := range comments {
			respComments[i] = &commentResponse{
				ID:        comm.ID,
				ParentID:  comm.ParentID,
				Text:      comm.Text,
				CreatedAt: comm.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		c.JSON(http.StatusOK, searchResponse{
			Comments: respComments,
			Total:    total,
			Page:     page,
			Limit:    limit,
		})
	}
}

// convertNodeToResponse - вспомогательная функция преобразования дерева из service в response
func convertNodeToResponse(node *service.CommentNode) *treeResponse {

	if node == nil {
		return nil
	}

	resp := &treeResponse{
		Comment: &commentResponse{
			ID:        node.Comment.ID,
			ParentID:  node.Comment.ParentID,
			Text:      node.Comment.Text,
			CreatedAt: node.Comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Children: make([]*treeResponse, len(node.Children)),
	}

	for i, child := range node.Children {
		resp.Children[i] = convertNodeToResponse(child)
	}

	return resp
}

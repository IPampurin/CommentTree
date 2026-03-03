package api

// Структуры для запросов/ответов API

// createCommentRequest тело запроса POST /comments
type createCommentRequest struct {
	ParentID *int   `json:"parent_id" binding:"omitempty"` // nil для корневых
	Text     string `json:"text"       binding:"required"`
}

// commentResponse базовый ответ с комментарием
type commentResponse struct {
	ID        int    `json:"id"`
	ParentID  *int   `json:"parent_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"` // RFC3339 для удобства
}

// treeResponse ответ с деревом комментариев
type treeResponse struct {
	Comment  *commentResponse `json:"comment"`
	Children []*treeResponse  `json:"children"`
}

// commentsListResponse ответ со списком комментариев и метаданными пагинации
type commentsListResponse struct {
	Comments []*commentResponse `json:"comments"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

// searchResponse ответ поиска
type searchResponse struct {
	Comments []*commentResponse `json:"comments"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

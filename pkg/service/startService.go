package service

import (
	"context"

	"github.com/IPampurin/CommentTree/pkg/db"
)

// Service инкапсулирует интерфейсный тип
type Service struct {
	storage db.CommentsMethods
}

// InitService возвращает экземпляр типа Service
func InitService(ctx context.Context, storage *db.DataBase) *Service {

	svc := &Service{
		storage, // *db.DataBase реализует CommentsMethods
	}

	return svc
}

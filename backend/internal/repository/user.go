package repository

import (
	"context"
	"football-backend/internal/model"
)

// UserRepository 定义用户数据访问接口
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*model.User, error)
}

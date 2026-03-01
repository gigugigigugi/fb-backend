package database

import (
	"context"

	"football-backend/internal/model"

	"gorm.io/gorm"
)

// Repository 定义了我们的通用数据库接口
type Repository interface {
	// 事务操作，传入一个可在闭包内执行 DB 操作的回调函数
	Transaction(ctx context.Context, fn func(txRepo Repository) error) error

	// 为后续具体查询保留的示例，可以根据真正的操作继续拓展
	// GetGormDB 返回底层的 *gorm.DB，如果在业务中仍然想要直接用 gorm
	// 最好避免，如果是为了跨不同种类的数据库(Pg vs DynamoDB)，
	// 推荐将具体的查询(如 FindUsers, FindMatch)写在 Repository 接口中。
	GetGormDB() *gorm.DB

	// --- 基础数据库操作 ---
	Create(ctx context.Context, value interface{}) error
	GetTeamByID(ctx context.Context, teamID uint) (*model.Team, error)

	// --- 领域级查询与防腐锁 ---
	GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error)
	HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error)
	CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error)
}

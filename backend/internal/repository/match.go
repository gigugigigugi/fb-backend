package repository

import (
	"context"
	"football-backend/internal/model"
	"time"
)

// MatchFilter 定义查询比赛列表的动态筛选条件
type MatchFilter struct {
	City      string
	StartTime time.Time
	Status    string
	Format    int
}

// MatchRepository 定义比赛赛程数据访问接口
type MatchRepository interface {
	CreateMatch(ctx context.Context, match *model.Match) error
	GetMatches(ctx context.Context, filter MatchFilter, offset, limit int) ([]*model.Match, int64, error)
	GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error)
	Transaction(ctx context.Context, fn func(txRepo MatchRepository) error) error
}

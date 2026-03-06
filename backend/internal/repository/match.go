package repository

import (
	"context"
	"football-backend/internal/model"
	"time"
)

// MatchFilter 定义比赛列表查询过滤条件。
type MatchFilter struct {
	City      string    // 城市过滤。
	StartTime time.Time // 开始时间下限过滤。
	Status    string    // 比赛状态过滤。
	Format    int       // 赛制过滤。
}

// MatchRepository 定义比赛数据访问接口。
type MatchRepository interface {
	CreateMatch(ctx context.Context, match *model.Match) error
	GetMatches(ctx context.Context, filter MatchFilter, offset, limit int) ([]*model.Match, int64, error)
	GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error)
	GetMatchByID(ctx context.Context, matchID uint) (*model.Match, error)
	GetCommentsByMatchID(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error)
	Transaction(ctx context.Context, fn func(txRepo MatchRepository) error) error
}

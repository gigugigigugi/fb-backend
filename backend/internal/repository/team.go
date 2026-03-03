package repository

import (
	"context"
	"football-backend/internal/model"
)

// TeamRepository 定义球队数据访问接口
type TeamRepository interface {
	CreateTeam(ctx context.Context, team *model.Team) error
	GetTeamByID(ctx context.Context, teamID uint) (*model.Team, error)
	IsTeamAdmin(ctx context.Context, teamID uint, userID uint) (bool, error)
}

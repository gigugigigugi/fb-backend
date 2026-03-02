package service

import (
	"context"
	"football-backend/common/database"
	"football-backend/internal/model"
)

type TeamService struct {
	repo database.Repository
}

func NewTeamService(repo database.Repository) *TeamService {
	return &TeamService{repo: repo}
}

// CreateTeam 创建球队
func (s *TeamService) CreateTeam(ctx context.Context, creatorID uint, teamName string) (*model.Team, error) {
	var team model.Team
	// 业务逻辑上提至 Service 控制事务包裹
	err := s.repo.Transaction(ctx, func(txRepo database.Repository) error {
		team = model.Team{Name: teamName, CaptainID: creatorID}
		if err := txRepo.Create(ctx, &team); err != nil {
			return err
		}
		member := model.TeamMember{TeamID: team.ID, UserID: creatorID, Role: "OWNER"}
		if err := txRepo.Create(ctx, &member); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetTeam 获取球队信息
func (s *TeamService) GetTeam(ctx context.Context, teamID uint) (*model.Team, error) {
	return s.repo.GetTeamByID(ctx, teamID)
}

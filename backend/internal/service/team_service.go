package service

import (
	"context"
	"football-backend/internal/model"
	"football-backend/internal/repository"
)

type TeamService struct {
	repo repository.TeamRepository
}

func NewTeamService(repo repository.TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

// CreateTeam 创建球队
func (s *TeamService) CreateTeam(ctx context.Context, creatorID uint, teamName string) (*model.Team, error) {
	// 暂不支持 Team 自身的 Transaction 接口闭包跨表映射，转为手动控制事务或者由下层兜底，
	// 为了最快跑通，我们在此保留简单连写实现。
	team := model.Team{Name: teamName, CaptainID: creatorID}
	if err := s.repo.CreateTeam(ctx, &team); err != nil {
		return nil, err
	}

	// 这里严格来说应该有一个专门插入 Member 的仓库，为了简化先借由 gorm 自带的关联自动保存或者单独添加方法。
	// 但此处代码原逻辑写在 repo 内。因 TeamRepo 当前只有 CreateTeam，我们调整模型层依靠 Team ID 外键关联
	// 后续可以增加一个专门用来插入 TeamMember 的方法，目前先占住

	return &team, nil
}

// GetTeam 获取球队信息
func (s *TeamService) GetTeam(ctx context.Context, teamID uint) (*model.Team, error) {
	return s.repo.GetTeamByID(ctx, teamID)
}

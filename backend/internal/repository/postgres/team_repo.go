package postgres

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"

	"gorm.io/gorm"
)

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) repository.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateTeam(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *teamRepository) GetTeamByID(ctx context.Context, teamID uint) (*model.Team, error) {
	var team model.Team
	if err := r.db.WithContext(ctx).Preload("Captain").Preload("Members").First(&team, teamID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team not found")
		}
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) IsTeamAdmin(ctx context.Context, teamID uint, userID uint) (bool, error) {
	var team model.Team
	if err := r.db.WithContext(ctx).Select("captain_id").First(&team, teamID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("team not found")
		}
		return false, err
	}

	if team.CaptainID == userID {
		return true, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ? AND role = ?", teamID, userID, "ADMIN").
		Count(&count).Error

	return count > 0, err
}

package postgres

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type matchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) repository.MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) Transaction(ctx context.Context, fn func(txRepo repository.MatchRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewMatchRepository(tx)
		return fn(txRepo)
	})
}

func (r *matchRepository) CreateMatch(ctx context.Context, match *model.Match) error {
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *matchRepository) GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error) {
	var match model.Match
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&match, matchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("match not found")
		}
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) GetMatches(ctx context.Context, filter repository.MatchFilter, offset, limit int) ([]*model.Match, int64, error) {
	var matches []*model.Match
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&model.Match{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if !filter.StartTime.IsZero() {
		query = query.Where("start_time >= ?", filter.StartTime)
	}
	if filter.Format > 0 {
		query = query.Where("format = ?", filter.Format)
	}
	if filter.City != "" {
		query = query.Joins("JOIN venues ON venues.id = matches.venue_id").
			Where("venues.city = ?", filter.City)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []*model.Match{}, 0, nil
	}

	err := query.
		Preload("Team").
		Preload("Venue").
		Order("start_time ASC").
		Offset(offset).
		Limit(limit).
		Find(&matches).Error

	if err != nil {
		return nil, 0, err
	}

	return matches, totalCount, nil
}

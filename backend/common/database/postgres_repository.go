package database

import (
	"context"
	"errors"

	"football-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresRepository 实现了 Repository 接口
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository constructor
func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{
		db: db,
	}
}

// Transaction 实现事务
func (p *PostgresRepository) Transaction(ctx context.Context, fn func(txRepo Repository) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewPostgresRepository(tx)
		return fn(txRepo)
	})
}

func (p *PostgresRepository) GetGormDB() *gorm.DB {
	return p.db
}

// Create 基础插入方法
func (p *PostgresRepository) Create(ctx context.Context, value interface{}) error {
	return p.db.WithContext(ctx).Create(value).Error
}

// GetMatchWithLock 排他锁获取比赛记录防止并发超卖
func (p *PostgresRepository) GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error) {
	var match model.Match
	if err := p.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&match, matchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("match not found")
		}
		return nil, err
	}
	return &match, nil
}

// HasUserBooked 检查用户是否已预约（不含取消记录）
func (p *PostgresRepository) HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error) {
	var existingCount int64
	err := p.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND user_id = ? AND status != ?", matchID, userID, "CANCELED").
		Count(&existingCount).Error
	return existingCount > 0, err
}

// CountConfirmedPlayers 查询已经成功加入本场比赛的活跃球员数
func (p *PostgresRepository) CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error) {
	var currentPlayers int64
	err := p.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND status = ?", matchID, "CONFIRMED").
		Count(&currentPlayers).Error
	return currentPlayers, err
}

// GetTeamByID 非事务性只读查询
func (p *PostgresRepository) GetTeamByID(ctx context.Context, teamID uint) (*model.Team, error) {
	var team model.Team
	// 直接使用 p.db 而无需 p.Transaction 闭包包裹，以保持轻量
	if err := p.db.WithContext(ctx).First(&team, teamID).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

package postgres

import (
	"context"
	"football-backend/internal/model"
	"football-backend/internal/repository"

	"gorm.io/gorm"
)

type venueRepository struct {
	db *gorm.DB
}

// NewVenueRepository 创建 VenueRepository 的 PostgreSQL 实现。
func NewVenueRepository(db *gorm.DB) repository.VenueRepository {
	return &venueRepository{db: db}
}

// GetRegionStats 按 prefecture/city 聚合统计场地数量。
func (r *venueRepository) GetRegionStats(ctx context.Context) ([]repository.VenueRegionRow, error) {
	rows := make([]repository.VenueRegionRow, 0, 12)
	err := r.db.WithContext(ctx).
		Model(&model.Venue{}).
		Select("prefecture, city, count(*) as venue_count").
		Where("prefecture <> '' AND city <> ''").
		Group("prefecture, city").
		Order("prefecture ASC, city ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetVenuesForMap 返回地图模式需要的场地点位数据（仅保留有经纬度的场地）。
func (r *venueRepository) GetVenuesForMap(ctx context.Context, filter repository.VenueMapFilter) ([]*model.Venue, error) {
	query := r.db.WithContext(ctx).Model(&model.Venue{}).
		Where("latitude <> 0 AND longitude <> 0")

	if filter.Prefecture != "" {
		query = query.Where("prefecture = ?", filter.Prefecture)
	}
	if filter.City != "" {
		query = query.Where("city = ?", filter.City)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	venues := make([]*model.Venue, 0, 12)
	err := query.
		Order("is_verified DESC, name ASC").
		Find(&venues).Error
	if err != nil {
		return nil, err
	}
	return venues, nil
}

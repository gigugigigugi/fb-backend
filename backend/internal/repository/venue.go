package repository

import (
	"context"
	"football-backend/internal/model"
)

// VenueRegionRow 表示按行政区聚合后的单行统计结果。
type VenueRegionRow struct {
	Prefecture string // 一级行政区，例如省/都道府县。
	City       string // 二级行政区，例如城市/区。
	VenueCount int64  // 当前行政区下的场地数量。
}

// VenueMapFilter 表示地图模式下的场地筛选参数。
type VenueMapFilter struct {
	Prefecture string // 可选：一级行政区过滤。
	City       string // 可选：城市过滤。
	Limit      int    // 返回条数上限。
}

// VenueRepository 定义场地查询的数据访问接口。
type VenueRepository interface {
	GetRegionStats(ctx context.Context) ([]VenueRegionRow, error)
	GetVenuesForMap(ctx context.Context, filter VenueMapFilter) ([]*model.Venue, error)
}

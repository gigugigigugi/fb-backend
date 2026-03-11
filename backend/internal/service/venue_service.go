package service

import (
	"context"
	"football-backend/internal/repository"
	"sort"
	"strings"
)

const (
	defaultVenueMapLimit = 200
	maxVenueMapLimit     = 500
)

// VenueRegionCity 表示一个城市及其场地数量。
type VenueRegionCity struct {
	City       string `json:"city"`        // 城市名称。
	VenueCount int64  `json:"venue_count"` // 城市下的场地数量。
}

// VenueRegionGroup 表示一个一级行政区及其城市列表。
type VenueRegionGroup struct {
	Prefecture string            `json:"prefecture"`  // 一级行政区名称。
	VenueCount int64             `json:"venue_count"` // 一级行政区下总场地数量。
	Cities     []VenueRegionCity `json:"cities"`      // 城市列表。
}

// VenueMapItem 表示地图接口返回的单个场地点位信息。
type VenueMapItem struct {
	ID         uint    `json:"id"`          // 场地 ID。
	Name       string  `json:"name"`        // 场地名称。
	Prefecture string  `json:"prefecture"`  // 一级行政区。
	City       string  `json:"city"`        // 城市。
	Address    string  `json:"address"`     // 详细地址。
	Latitude   float64 `json:"latitude"`    // 纬度。
	Longitude  float64 `json:"longitude"`   // 经度。
	IsVerified bool    `json:"is_verified"` // 是否已认证。
}

// VenueService 负责场地发现相关业务编排。
type VenueService struct {
	repo repository.VenueRepository
}

// NewVenueService 创建 VenueService。
func NewVenueService(repo repository.VenueRepository) *VenueService {
	return &VenueService{repo: repo}
}

// GetRegions 返回按 prefecture -> city 分组的场地统计结构。
func (s *VenueService) GetRegions(ctx context.Context) ([]VenueRegionGroup, error) {
	rows, err := s.repo.GetRegionStats(ctx)
	if err != nil {
		return nil, err
	}

	groups := make(map[string]*VenueRegionGroup, len(rows))
	for _, row := range rows {
		prefecture := strings.TrimSpace(row.Prefecture)
		city := strings.TrimSpace(row.City)
		if prefecture == "" || city == "" {
			continue
		}

		group, exists := groups[prefecture]
		if !exists {
			group = &VenueRegionGroup{
				Prefecture: prefecture,
				VenueCount: 0,
				Cities:     make([]VenueRegionCity, 0, 4),
			}
			groups[prefecture] = group
		}

		group.Cities = append(group.Cities, VenueRegionCity{
			City:       city,
			VenueCount: row.VenueCount,
		})
		group.VenueCount += row.VenueCount
	}

	result := make([]VenueRegionGroup, 0, len(groups))
	for _, group := range groups {
		sort.Slice(group.Cities, func(i, j int) bool {
			return group.Cities[i].City < group.Cities[j].City
		})
		result = append(result, *group)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Prefecture < result[j].Prefecture
	})

	return result, nil
}

// GetMapVenues 返回地图展示需要的场地点位列表。
func (s *VenueService) GetMapVenues(ctx context.Context, prefecture, city string, limit int) ([]VenueMapItem, error) {
	if limit <= 0 {
		limit = defaultVenueMapLimit
	}
	if limit > maxVenueMapLimit {
		limit = maxVenueMapLimit
	}

	filter := repository.VenueMapFilter{
		Prefecture: strings.TrimSpace(prefecture),
		City:       strings.TrimSpace(city),
		Limit:      limit,
	}
	venues, err := s.repo.GetVenuesForMap(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]VenueMapItem, 0, len(venues))
	for _, v := range venues {
		items = append(items, VenueMapItem{
			ID:         v.ID,
			Name:       v.Name,
			Prefecture: v.Prefecture,
			City:       v.City,
			Address:    v.Address,
			Latitude:   v.Latitude,
			Longitude:  v.Longitude,
			IsVerified: v.IsVerified,
		})
	}
	return items, nil
}

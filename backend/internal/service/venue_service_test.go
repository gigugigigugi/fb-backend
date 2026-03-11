package service

import (
	"context"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"testing"
)

func TestVenueServiceGetRegions(t *testing.T) {
	svc := NewVenueService(&mockVenueRepo{
		getRegionStatsFn: func(ctx context.Context) ([]repository.VenueRegionRow, error) {
			return []repository.VenueRegionRow{
				{Prefecture: "Tokyo", City: "Shibuya", VenueCount: 2},
				{Prefecture: "Tokyo", City: "Koto", VenueCount: 3},
				{Prefecture: "Kanagawa", City: "Yokohama", VenueCount: 1},
			}, nil
		},
	})

	regions, err := svc.GetRegions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(regions) != 2 {
		t.Fatalf("expected 2 prefectures, got %d", len(regions))
	}
	if regions[0].Prefecture != "Kanagawa" || regions[1].Prefecture != "Tokyo" {
		t.Fatalf("unexpected order: %#v", regions)
	}
	if regions[1].VenueCount != 5 {
		t.Fatalf("expected Tokyo venue_count=5, got %d", regions[1].VenueCount)
	}
}

func TestVenueServiceGetMapVenuesLimitAndTransform(t *testing.T) {
	svc := NewVenueService(&mockVenueRepo{
		getMapVenuesFn: func(ctx context.Context, filter repository.VenueMapFilter) ([]*model.Venue, error) {
			if filter.Limit != 500 {
				t.Fatalf("expected clamped limit=500, got %d", filter.Limit)
			}
			if filter.Prefecture != "Tokyo" || filter.City != "Shibuya" {
				t.Fatalf("unexpected filter: %#v", filter)
			}
			return []*model.Venue{
				{
					ID:         1,
					Name:       "A",
					Prefecture: "Tokyo",
					City:       "Shibuya",
					Address:    "Addr",
					Latitude:   35.1,
					Longitude:  139.1,
					IsVerified: true,
				},
			}, nil
		},
	})

	items, err := svc.GetMapVenues(context.Background(), " Tokyo ", " Shibuya ", 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Name != "A" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

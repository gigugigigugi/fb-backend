package handler

import (
	"context"
	"encoding/json"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"football-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVenueHandlerGetRegions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubVenueRepo{
		getRegionStatsFn: func(ctx context.Context) ([]repository.VenueRegionRow, error) {
			return []repository.VenueRegionRow{
				{Prefecture: "Tokyo", City: "Shibuya", VenueCount: 2},
			}, nil
		},
	}
	h := NewVenueHandler(service.NewVenueService(repo))

	r := gin.New()
	r.GET("/venues/regions", h.GetRegions)

	req := httptest.NewRequest(http.MethodGet, "/venues/regions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if int(got["code"].(float64)) != 0 {
		t.Fatalf("expected code=0, got %v", got["code"])
	}
}

func TestVenueHandlerGetMap(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubVenueRepo{
		getMapVenuesFn: func(ctx context.Context, filter repository.VenueMapFilter) ([]*model.Venue, error) {
			return []*model.Venue{
				{
					ID:         1,
					Name:       "Field A",
					Prefecture: "Tokyo",
					City:       "Shibuya",
					Latitude:   35.1,
					Longitude:  139.1,
					IsVerified: true,
				},
			}, nil
		},
	}
	h := NewVenueHandler(service.NewVenueService(repo))

	r := gin.New()
	r.GET("/venues/map", h.GetMap)

	req := httptest.NewRequest(http.MethodGet, "/venues/map?prefecture=Tokyo&city=Shibuya&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if int(got["code"].(float64)) != 0 {
		t.Fatalf("expected code=0, got %v", got["code"])
	}
}

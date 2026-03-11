package handler

import (
	"context"
	"encoding/json"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"football-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMatchHandlerSettleMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, TeamID: 8}, nil
			},
		}
		bookingRepo := &stubBookingRepo{
			settleMatchBookingsFn: func(ctx context.Context, matchID uint, paymentStatus string, bookingIDs []uint) (int64, error) {
				if paymentStatus != "PAID" {
					t.Fatalf("expected PAID, got %s", paymentStatus)
				}
				return 2, nil
			},
		}
		teamRepo := &stubTeamRepo{
			isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
				return true, nil
			},
		}

		h := NewMatchHandler(service.NewMatchService(matchRepo, bookingRepo, teamRepo, &stubUserRepo{}, nil))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.POST("/matches/:id/settlement", h.SettleMatch)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/settlement", strings.NewReader(`{"payment_status":"paid","booking_ids":[1,2]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
		}
		var got map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		data := got["data"].(map[string]interface{})
		if int(data["updated_count"].(float64)) != 2 {
			t.Fatalf("expected updated_count=2, got %v", data["updated_count"])
		}
	})

	t.Run("unauthorized when userID missing", func(t *testing.T) {
		h := NewMatchHandler(service.NewMatchService(&stubMatchRepo{}, &stubBookingRepo{}, &stubTeamRepo{}, &stubUserRepo{}, nil))
		r := gin.New()
		r.POST("/matches/:id/settlement", h.SettleMatch)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/settlement", strings.NewReader(`{"payment_status":"PAID"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, TeamID: 8}, nil
			},
		}
		teamRepo := &stubTeamRepo{
			isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
				return false, nil
			},
		}
		h := NewMatchHandler(service.NewMatchService(matchRepo, &stubBookingRepo{}, teamRepo, &stubUserRepo{}, nil))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.POST("/matches/:id/settlement", h.SettleMatch)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/settlement", strings.NewReader(`{"payment_status":"PAID"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("invalid payment status", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, TeamID: 8}, nil
			},
		}
		teamRepo := &stubTeamRepo{
			isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
				return true, nil
			},
		}
		h := NewMatchHandler(service.NewMatchService(matchRepo, &stubBookingRepo{}, teamRepo, &stubUserRepo{}, nil))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.POST("/matches/:id/settlement", h.SettleMatch)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/settlement", strings.NewReader(`{"payment_status":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestMatchHandlerAssignSubTeams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, TeamID: 8}, nil
			},
		}
		bookingRepo := &stubBookingRepo{
			assignSubTeamsFn: func(ctx context.Context, matchID uint, assignments []repository.SubTeamAssignment) error {
				if len(assignments) != 2 {
					t.Fatalf("expected 2 assignments, got %d", len(assignments))
				}
				return nil
			},
		}
		teamRepo := &stubTeamRepo{
			isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
				return true, nil
			},
		}
		h := NewMatchHandler(service.NewMatchService(matchRepo, bookingRepo, teamRepo, &stubUserRepo{}, nil))

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.POST("/matches/:id/subteams", h.AssignSubTeams)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/subteams", strings.NewReader(`{"assignments":[{"booking_id":1,"sub_team":"A"},{"booking_id":2,"sub_team":"B"}]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("unauthorized when userID missing", func(t *testing.T) {
		h := NewMatchHandler(service.NewMatchService(&stubMatchRepo{}, &stubBookingRepo{}, &stubTeamRepo{}, &stubUserRepo{}, nil))
		r := gin.New()
		r.POST("/matches/:id/subteams", h.AssignSubTeams)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/subteams", strings.NewReader(`{"assignments":[{"booking_id":1,"sub_team":"A"}]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid assignments", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, TeamID: 8}, nil
			},
		}
		teamRepo := &stubTeamRepo{
			isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
				return true, nil
			},
		}
		h := NewMatchHandler(service.NewMatchService(matchRepo, &stubBookingRepo{}, teamRepo, &stubUserRepo{}, nil))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.POST("/matches/:id/subteams", h.AssignSubTeams)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/subteams", strings.NewReader(`{"assignments":[{"booking_id":1,"sub_team":""}]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("match not found", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return nil, errors.New("match not found")
			},
		}
		h := NewMatchHandler(service.NewMatchService(matchRepo, &stubBookingRepo{}, &stubTeamRepo{}, &stubUserRepo{}, nil))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.POST("/matches/:id/subteams", h.AssignSubTeams)

		req := httptest.NewRequest(http.MethodPost, "/matches/100/subteams", strings.NewReader(`{"assignments":[{"booking_id":1,"sub_team":"A"}]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

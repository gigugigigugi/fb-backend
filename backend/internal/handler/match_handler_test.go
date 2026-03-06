package handler

import (
	"context"
	"encoding/json"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMatchHandlerGetMatchDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, Status: "RECRUITING"}, nil
			},
			getCommentsByMatchIDFn: func(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error) {
				return []*model.Comment{{
					ID:        1,
					UserID:    7,
					Content:   "hello",
					CreatedAt: time.Now(),
					User:      &model.User{ID: 7, Nickname: "Alice", Avatar: "a.png"},
				}}, nil
			},
		}
		bookingRepo := &stubBookingRepo{
			getBookingsByMatchIDFn: func(ctx context.Context, matchID uint) ([]*model.Booking, error) {
				return []*model.Booking{{
					ID:        11,
					UserID:    7,
					GuestName: "",
					Status:    "CONFIRMED",
					User:      &model.User{ID: 7, Nickname: "Alice", Avatar: "a.png"},
				}}, nil
			},
		}

		svc := service.NewMatchService(matchRepo, bookingRepo, &stubTeamRepo{}, &stubUserRepo{}, nil)
		h := NewMatchHandler(svc)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.GET("/matches/:id/details", h.GetMatchDetails)

		req := httptest.NewRequest(http.MethodGet, "/matches/100/details", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
		}
		var got map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal response failed: %v", err)
		}
		if int(got["code"].(float64)) != 0 {
			t.Fatalf("expected code=0, got: %v", got["code"])
		}
		data := got["data"].(map[string]interface{})
		if data["user_status"].(string) != "JOINED" {
			t.Fatalf("expected user_status JOINED, got: %v", data["user_status"])
		}
	})

	t.Run("bad match id", func(t *testing.T) {
		h := NewMatchHandler(service.NewMatchService(&stubMatchRepo{}, &stubBookingRepo{}, &stubTeamRepo{}, &stubUserRepo{}, nil))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.GET("/matches/:id/details", h.GetMatchDetails)

		req := httptest.NewRequest(http.MethodGet, "/matches/abc/details", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized when userID missing", func(t *testing.T) {
		h := NewMatchHandler(service.NewMatchService(&stubMatchRepo{}, &stubBookingRepo{}, &stubTeamRepo{}, &stubUserRepo{}, nil))
		r := gin.New()
		r.GET("/matches/:id/details", h.GetMatchDetails)

		req := httptest.NewRequest(http.MethodGet, "/matches/100/details", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		matchRepo := &stubMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return nil, errors.New("match not found")
			},
		}
		svc := service.NewMatchService(matchRepo, &stubBookingRepo{}, &stubTeamRepo{}, &stubUserRepo{}, nil)
		h := NewMatchHandler(svc)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.GET("/matches/:id/details", h.GetMatchDetails)

		req := httptest.NewRequest(http.MethodGet, "/matches/999/details", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

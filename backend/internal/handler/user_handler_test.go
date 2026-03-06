package handler

import (
	"context"
	"encoding/json"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserHandlerGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		repo := &stubUserRepo{
			getUserByIDFn: func(_ context.Context, id uint) (*model.User, error) {
				return &model.User{ID: id, Nickname: "Alice", Email: "alice@example.com"}, nil
			},
		}
		h := NewUserHandler(service.NewUserService(repo))

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.GET("/users/me", h.GetMe)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
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
			t.Fatalf("expected code=0, got: %v", got["code"])
		}
		data := got["data"].(map[string]interface{})
		if int(data["id"].(float64)) != 7 {
			t.Fatalf("expected data.id=7, got: %v", data["id"])
		}
	})

	t.Run("unauthorized when userID missing", func(t *testing.T) {
		h := NewUserHandler(service.NewUserService(&stubUserRepo{}))
		r := gin.New()
		r.GET("/users/me", h.GetMe)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &stubUserRepo{
			getUserByIDFn: func(_ context.Context, id uint) (*model.User, error) {
				return nil, errors.New("user not found")
			},
		}
		h := NewUserHandler(service.NewUserService(repo))

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(8))
			c.Next()
		})
		r.GET("/users/me", h.GetMe)

		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestUserHandlerUpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		repo := &stubUserRepo{
			updateUserProfileFn: func(_ context.Context, userID uint, nickname *string, avatar *string) (*model.User, error) {
				if userID != 7 {
					t.Fatalf("unexpected userID=%d", userID)
				}
				if nickname == nil || *nickname != "Alice" {
					t.Fatalf("expected trimmed nickname Alice, got %#v", nickname)
				}
				if avatar == nil || *avatar != "https://img.example.com/a.png" {
					t.Fatalf("expected trimmed avatar, got %#v", avatar)
				}
				return &model.User{ID: userID, Nickname: *nickname, Avatar: *avatar}, nil
			},
		}
		h := NewUserHandler(service.NewUserService(repo))

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.PUT("/users/me", h.UpdateMe)

		body := `{"nickname":"  Alice  ","avatar":" https://img.example.com/a.png "}`
		req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(body))
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
		if int(got["code"].(float64)) != 0 {
			t.Fatalf("expected code=0, got: %v", got["code"])
		}
	})

	t.Run("invalid json type", func(t *testing.T) {
		h := NewUserHandler(service.NewUserService(&stubUserRepo{}))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.PUT("/users/me", h.UpdateMe)

		body := `{"nickname":123}`
		req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("service validation error", func(t *testing.T) {
		h := NewUserHandler(service.NewUserService(&stubUserRepo{}))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("userID", uint(7))
			c.Next()
		})
		r.PUT("/users/me", h.UpdateMe)

		body := `{"nickname":"a"}`
		req := httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		var got map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &got)
		if int(got["code"].(float64)) != 400 {
			t.Fatalf("expected code=400, got: %v", got["code"])
		}
	})
}

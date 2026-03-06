package service

import (
	"context"
	"strings"
	"testing"

	"football-backend/internal/model"
)

func TestUserServiceUpdateMeValidation(t *testing.T) {
	svc := NewUserService(&mockUserRepo{})

	t.Run("empty payload", func(t *testing.T) {
		_, err := svc.UpdateMe(context.Background(), 1, UpdateMeInput{})
		if err == nil || err.Error() != "nothing to update" {
			t.Fatalf("expected nothing to update error, got: %v", err)
		}
	})

	t.Run("nickname too short", func(t *testing.T) {
		nickname := "a"
		_, err := svc.UpdateMe(context.Background(), 1, UpdateMeInput{Nickname: &nickname})
		if err == nil || !strings.Contains(err.Error(), "nickname length") {
			t.Fatalf("expected nickname length error, got: %v", err)
		}
	})

	t.Run("nickname too long", func(t *testing.T) {
		nickname := strings.Repeat("n", 51)
		_, err := svc.UpdateMe(context.Background(), 1, UpdateMeInput{Nickname: &nickname})
		if err == nil || !strings.Contains(err.Error(), "nickname length") {
			t.Fatalf("expected nickname length error, got: %v", err)
		}
	})

	t.Run("avatar too long", func(t *testing.T) {
		avatar := strings.Repeat("a", 256)
		_, err := svc.UpdateMe(context.Background(), 1, UpdateMeInput{Avatar: &avatar})
		if err == nil || !strings.Contains(err.Error(), "avatar url is too long") {
			t.Fatalf("expected avatar length error, got: %v", err)
		}
	})
}

func TestUserServiceUpdateMeSuccessTrimmed(t *testing.T) {
	nickname := "  Alice  "
	avatar := " https://img.example/avatar.png "

	repo := &mockUserRepo{
		updateUserProfileFn: func(ctx context.Context, userID uint, gotNickname *string, gotAvatar *string) (*model.User, error) {
			if userID != 7 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			if gotNickname == nil || *gotNickname != "Alice" {
				t.Fatalf("nickname should be trimmed to Alice, got: %#v", gotNickname)
			}
			if gotAvatar == nil || *gotAvatar != "https://img.example/avatar.png" {
				t.Fatalf("avatar should be trimmed, got: %#v", gotAvatar)
			}
			return &model.User{ID: userID, Nickname: *gotNickname, Avatar: *gotAvatar}, nil
		},
	}

	svc := NewUserService(repo)
	got, err := svc.UpdateMe(context.Background(), 7, UpdateMeInput{
		Nickname: &nickname,
		Avatar:   &avatar,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Nickname != "Alice" || got.Avatar != "https://img.example/avatar.png" {
		t.Fatalf("unexpected updated user: %#v", got)
	}
}

func TestUserServiceGetMe(t *testing.T) {
	repo := &mockUserRepo{
		getUserByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
			return &model.User{ID: id, Email: "u@example.com", Nickname: "tester"}, nil
		},
	}

	svc := NewUserService(repo)
	got, err := svc.GetMe(context.Background(), 11)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != 11 {
		t.Fatalf("unexpected user: %#v", got)
	}
}

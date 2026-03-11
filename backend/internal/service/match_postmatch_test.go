package service

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"testing"
)

func TestMatchServiceSettleMatchPermissionAndStatus(t *testing.T) {
	t.Run("reject invalid payment status", func(t *testing.T) {
		svc := NewMatchService(&mockMatchRepo{}, &mockBookingRepo{}, &mockTeamRepo{}, &mockUserRepo{}, nil)
		_, err := svc.SettleMatch(context.Background(), 1, 2, "INVALID", nil)
		if !errors.Is(err, ErrInvalidPaymentStatus) {
			t.Fatalf("expected ErrInvalidPaymentStatus, got %v", err)
		}
	})

	t.Run("captain or admin can settle", func(t *testing.T) {
		called := false
		svc := NewMatchService(
			&mockMatchRepo{
				getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
					return &model.Match{ID: matchID, TeamID: 9}, nil
				},
			},
			&mockBookingRepo{
				settleMatchBookingsFn: func(ctx context.Context, matchID uint, paymentStatus string, bookingIDs []uint) (int64, error) {
					called = true
					if paymentStatus != "PAID" {
						t.Fatalf("expected payment status PAID, got %s", paymentStatus)
					}
					if len(bookingIDs) != 2 || bookingIDs[0] != 1 || bookingIDs[1] != 2 {
						t.Fatalf("expected deduped booking IDs [1,2], got %#v", bookingIDs)
					}
					return 2, nil
				},
			},
			&mockTeamRepo{
				isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
					return true, nil
				},
			},
			&mockUserRepo{},
			nil,
		)

		updated, err := svc.SettleMatch(context.Background(), 11, 7, "paid", []uint{1, 1, 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected SettleMatchBookings to be called")
		}
		if updated != 2 {
			t.Fatalf("expected updated=2, got %d", updated)
		}
	})

	t.Run("member cannot settle", func(t *testing.T) {
		called := false
		svc := NewMatchService(
			&mockMatchRepo{
				getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
					return &model.Match{ID: matchID, TeamID: 9}, nil
				},
			},
			&mockBookingRepo{
				settleMatchBookingsFn: func(ctx context.Context, matchID uint, paymentStatus string, bookingIDs []uint) (int64, error) {
					called = true
					return 0, nil
				},
			},
			&mockTeamRepo{
				isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
					return false, nil
				},
			},
			&mockUserRepo{},
			nil,
		)

		_, err := svc.SettleMatch(context.Background(), 11, 7, "PAID", nil)
		if !errors.Is(err, ErrMatchManageForbidden) {
			t.Fatalf("expected ErrMatchManageForbidden, got %v", err)
		}
		if called {
			t.Fatal("SettleMatchBookings should not be called when forbidden")
		}
	})
}

func TestMatchServiceAssignSubTeamsPermissionAndValidation(t *testing.T) {
	baseMatchRepo := &mockMatchRepo{
		getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
			return &model.Match{ID: matchID, TeamID: 3}, nil
		},
	}

	t.Run("invalid assignments should fail", func(t *testing.T) {
		svc := NewMatchService(
			baseMatchRepo,
			&mockBookingRepo{},
			&mockTeamRepo{
				isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
					return true, nil
				},
			},
			&mockUserRepo{},
			nil,
		)

		err := svc.AssignMatchSubTeams(context.Background(), 1, 9, []SubTeamAssignment{
			{BookingID: 1, SubTeam: ""},
		})
		if !errors.Is(err, ErrInvalidSubTeamAssignments) {
			t.Fatalf("expected ErrInvalidSubTeamAssignments, got %v", err)
		}
	})

	t.Run("duplicate booking id should fail", func(t *testing.T) {
		svc := NewMatchService(
			baseMatchRepo,
			&mockBookingRepo{},
			&mockTeamRepo{
				isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
					return true, nil
				},
			},
			&mockUserRepo{},
			nil,
		)

		err := svc.AssignMatchSubTeams(context.Background(), 1, 9, []SubTeamAssignment{
			{BookingID: 1, SubTeam: "A"},
			{BookingID: 1, SubTeam: "B"},
		})
		if !errors.Is(err, ErrInvalidSubTeamAssignments) {
			t.Fatalf("expected ErrInvalidSubTeamAssignments, got %v", err)
		}
	})

	t.Run("admin can assign subteams", func(t *testing.T) {
		called := false
		svc := NewMatchService(
			baseMatchRepo,
			&mockBookingRepo{
				assignSubTeamsFn: func(ctx context.Context, matchID uint, assignments []repository.SubTeamAssignment) error {
					called = true
					if len(assignments) != 2 {
						t.Fatalf("expected 2 assignments, got %d", len(assignments))
					}
					if assignments[0].SubTeam != "A" || assignments[1].SubTeam != "B" {
						t.Fatalf("unexpected assignments: %#v", assignments)
					}
					return nil
				},
			},
			&mockTeamRepo{
				isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
					return true, nil
				},
			},
			&mockUserRepo{},
			nil,
		)

		err := svc.AssignMatchSubTeams(context.Background(), 1, 9, []SubTeamAssignment{
			{BookingID: 10, SubTeam: " A "},
			{BookingID: 11, SubTeam: "B"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected AssignSubTeams to be called")
		}
	})

	t.Run("member cannot assign subteams", func(t *testing.T) {
		svc := NewMatchService(
			baseMatchRepo,
			&mockBookingRepo{},
			&mockTeamRepo{
				isTeamAdminFn: func(ctx context.Context, teamID uint, userID uint) (bool, error) {
					return false, nil
				},
			},
			&mockUserRepo{},
			nil,
		)

		err := svc.AssignMatchSubTeams(context.Background(), 1, 9, []SubTeamAssignment{
			{BookingID: 10, SubTeam: "A"},
		})
		if !errors.Is(err, ErrMatchManageForbidden) {
			t.Fatalf("expected ErrMatchManageForbidden, got %v", err)
		}
	})

	t.Run("match not found", func(t *testing.T) {
		svc := NewMatchService(
			&mockMatchRepo{
				getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
					return nil, errors.New("match not found")
				},
			},
			&mockBookingRepo{},
			&mockTeamRepo{},
			&mockUserRepo{},
			nil,
		)

		err := svc.AssignMatchSubTeams(context.Background(), 99, 1, []SubTeamAssignment{
			{BookingID: 10, SubTeam: "A"},
		})
		if !errors.Is(err, ErrMatchNotFound) {
			t.Fatalf("expected ErrMatchNotFound, got %v", err)
		}
	})
}

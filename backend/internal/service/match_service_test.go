package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"football-backend/common/notification"
	"football-backend/internal/model"
)

type countingNotifier struct {
	channel notification.Channel
	sentCh  chan uint
}

func (n *countingNotifier) Channel() notification.Channel {
	return n.channel
}

func (n *countingNotifier) Send(ctx context.Context, recipient notification.Recipient, msg notification.Message) error {
	n.sentCh <- recipient.UserID
	return nil
}

func TestMatchServiceGetMatchDetailsUserStatus(t *testing.T) {
	cases := []struct {
		name     string
		userID   uint
		bookings []*model.Booking
		want     string
	}{
		{
			name:     "not joined",
			userID:   9,
			bookings: []*model.Booking{{UserID: 3, Status: "CONFIRMED"}},
			want:     "NOT_JOINED",
		},
		{
			name:     "waiting",
			userID:   9,
			bookings: []*model.Booking{{UserID: 9, Status: "WAITING"}},
			want:     "WAITING",
		},
		{
			name:     "canceled only",
			userID:   9,
			bookings: []*model.Booking{{UserID: 9, Status: "CANCELED"}},
			want:     "CANCELED",
		},
		{
			name:   "confirmed has highest priority",
			userID: 9,
			bookings: []*model.Booking{
				{UserID: 9, Status: "WAITING"},
				{UserID: 9, Status: "CONFIRMED"},
			},
			want: "JOINED",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMatchService(
				&mockMatchRepo{
					getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
						return &model.Match{ID: matchID, Status: "RECRUITING"}, nil
					},
					getCommentsByMatchIDFn: func(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error) {
						return []*model.Comment{}, nil
					},
				},
				&mockBookingRepo{
					getBookingsByMatchIDFn: func(ctx context.Context, matchID uint) ([]*model.Booking, error) {
						return tc.bookings, nil
					},
				},
				&mockTeamRepo{},
				&mockUserRepo{},
				nil,
			)

			resp, err := svc.GetMatchDetails(context.Background(), 1, tc.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.UserStatus != tc.want {
				t.Fatalf("expected user status %s, got %s", tc.want, resp.UserStatus)
			}
		})
	}
}

func TestMatchServiceGetMatchDetailsRosterAndComments(t *testing.T) {
	now := time.Now().UTC()

	bookings := []*model.Booking{
		{
			ID:        1,
			UserID:    10,
			GuestName: "",
			Status:    "CONFIRMED",
			User:      &model.User{ID: 10, Nickname: "Alice", Avatar: "alice.png"},
		},
		{
			ID:        2,
			UserID:    11,
			GuestName: "Friend",
			Status:    "WAITING",
			User:      &model.User{ID: 11, Nickname: "Bob", Avatar: "bob.png"},
		},
	}

	comments := []*model.Comment{
		{
			ID:        5,
			UserID:    10,
			Content:   "See you on field",
			CreatedAt: now,
			User:      &model.User{ID: 10, Nickname: "Alice", Avatar: "alice.png"},
		},
	}

	svc := NewMatchService(
		&mockMatchRepo{
			getMatchByIDFn: func(ctx context.Context, matchID uint) (*model.Match, error) {
				return &model.Match{ID: matchID, Status: "RECRUITING"}, nil
			},
			getCommentsByMatchIDFn: func(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error) {
				if limit != 50 {
					t.Fatalf("expected comments limit=50, got %d", limit)
				}
				return comments, nil
			},
		},
		&mockBookingRepo{
			getBookingsByMatchIDFn: func(ctx context.Context, matchID uint) ([]*model.Booking, error) {
				return bookings, nil
			},
		},
		&mockTeamRepo{},
		&mockUserRepo{},
		nil,
	)

	resp, err := svc.GetMatchDetails(context.Background(), 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.UserStatus != "JOINED" {
		t.Fatalf("expected user status JOINED, got %s", resp.UserStatus)
	}
	if len(resp.Roster.Confirmed) != 1 || len(resp.Roster.Waiting) != 1 {
		t.Fatalf("unexpected roster groups: %#v", resp.Roster)
	}
	if resp.Roster.Confirmed[0].Nickname != "Alice" || resp.Roster.Waiting[0].GuestName != "Friend" {
		t.Fatalf("unexpected roster content: %#v", resp.Roster)
	}
	if len(resp.Comments) != 1 || resp.Comments[0].Nickname != "Alice" || resp.Comments[0].Content != "See you on field" {
		t.Fatalf("unexpected comments: %#v", resp.Comments)
	}
}

func TestMatchServiceCancelBookingNotifyMax10(t *testing.T) {
	waitingIDs := make([]uint, 12)
	for i := range waitingIDs {
		waitingIDs[i] = uint(i + 1)
	}

	emailNotifier := &countingNotifier{
		channel: notification.ChannelEmail,
		sentCh:  make(chan uint, 20),
	}
	dispatcher := notification.NewDispatcher(64)
	dispatcher.RegisterNotifier(emailNotifier)

	svc := NewMatchService(
		&mockMatchRepo{},
		&mockBookingRepo{
			cancelBookingTransactionFn: func(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error) {
				return 88, waitingIDs, nil
			},
		},
		&mockTeamRepo{},
		&mockUserRepo{
			getUserByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
				email := fmt.Sprintf("u%d@example.com", id)
				return &model.User{
					ID:            id,
					Email:         email,
					EmailVerified: true,
				}, nil
			},
		},
		dispatcher,
	)

	if err := svc.CancelBooking(context.Background(), 10, 999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	received := make([]uint, 0, 10)
	deadline := time.After(2 * time.Second)
	for len(received) < 10 {
		select {
		case id := <-emailNotifier.sentCh:
			received = append(received, id)
		case <-deadline:
			t.Fatalf("expected 10 notifications, only got %d (%v)", len(received), received)
		}
	}

	for i := 0; i < 10; i++ {
		want := uint(i + 1)
		if received[i] != want {
			t.Fatalf("expected notified user at index %d is %d, got %d", i, want, received[i])
		}
	}

	select {
	case extra := <-emailNotifier.sentCh:
		t.Fatalf("should notify at most 10 users, but got extra user id=%d", extra)
	case <-time.After(300 * time.Millisecond):
		// pass
	}
}

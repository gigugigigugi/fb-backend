package handler

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
)

func panicUnexpectedHandlerCall(name string) {
	panic("unexpected call: " + name)
}

type stubUserRepo struct {
	getUserByIDFn       func(ctx context.Context, id uint) (*model.User, error)
	updateUserProfileFn func(ctx context.Context, userID uint, nickname *string, avatar *string) (*model.User, error)
}

var _ repository.UserRepository = (*stubUserRepo)(nil)

func (s *stubUserRepo) CreateUser(ctx context.Context, user *model.User) error {
	panicUnexpectedHandlerCall("CreateUser")
	return nil
}

func (s *stubUserRepo) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	if s.getUserByIDFn == nil {
		return nil, errors.New("user not found")
	}
	return s.getUserByIDFn(ctx, id)
}

func (s *stubUserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	panicUnexpectedHandlerCall("GetUserByEmail")
	return nil, nil
}

func (s *stubUserRepo) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	panicUnexpectedHandlerCall("GetUserByPhone")
	return nil, nil
}

func (s *stubUserRepo) GetUserByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	panicUnexpectedHandlerCall("GetUserByGoogleID")
	return nil, nil
}

func (s *stubUserRepo) UpdateUserProfile(ctx context.Context, userID uint, nickname *string, avatar *string) (*model.User, error) {
	if s.updateUserProfileFn == nil {
		panicUnexpectedHandlerCall("UpdateUserProfile")
	}
	return s.updateUserProfileFn(ctx, userID, nickname, avatar)
}

func (s *stubUserRepo) UpdateEmailVerified(ctx context.Context, userID uint, verified bool) error {
	panicUnexpectedHandlerCall("UpdateEmailVerified")
	return nil
}

func (s *stubUserRepo) UpdatePhoneVerified(ctx context.Context, userID uint, phone string, verified bool) error {
	panicUnexpectedHandlerCall("UpdatePhoneVerified")
	return nil
}

type stubMatchRepo struct {
	getMatchByIDFn         func(ctx context.Context, matchID uint) (*model.Match, error)
	getCommentsByMatchIDFn func(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error)
}

var _ repository.MatchRepository = (*stubMatchRepo)(nil)

func (s *stubMatchRepo) CreateMatch(ctx context.Context, match *model.Match) error {
	panicUnexpectedHandlerCall("CreateMatch")
	return nil
}

func (s *stubMatchRepo) GetMatches(ctx context.Context, filter repository.MatchFilter, offset, limit int) ([]*model.Match, int64, error) {
	panicUnexpectedHandlerCall("GetMatches")
	return nil, 0, nil
}

func (s *stubMatchRepo) GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error) {
	panicUnexpectedHandlerCall("GetMatchWithLock")
	return nil, nil
}

func (s *stubMatchRepo) GetMatchByID(ctx context.Context, matchID uint) (*model.Match, error) {
	if s.getMatchByIDFn == nil {
		panicUnexpectedHandlerCall("GetMatchByID")
	}
	return s.getMatchByIDFn(ctx, matchID)
}

func (s *stubMatchRepo) GetCommentsByMatchID(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error) {
	if s.getCommentsByMatchIDFn == nil {
		return []*model.Comment{}, nil
	}
	return s.getCommentsByMatchIDFn(ctx, matchID, limit)
}

func (s *stubMatchRepo) Transaction(ctx context.Context, fn func(txRepo repository.MatchRepository) error) error {
	return fn(s)
}

type stubBookingRepo struct {
	getBookingsByMatchIDFn func(ctx context.Context, matchID uint) ([]*model.Booking, error)
}

var _ repository.BookingRepository = (*stubBookingRepo)(nil)

func (s *stubBookingRepo) CreateBooking(ctx context.Context, booking *model.Booking) error {
	panicUnexpectedHandlerCall("CreateBooking")
	return nil
}

func (s *stubBookingRepo) HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error) {
	panicUnexpectedHandlerCall("HasUserBooked")
	return false, nil
}

func (s *stubBookingRepo) CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error) {
	panicUnexpectedHandlerCall("CountConfirmedPlayers")
	return 0, nil
}

func (s *stubBookingRepo) CountWaitingPlayers(ctx context.Context, matchID uint) (int64, error) {
	panicUnexpectedHandlerCall("CountWaitingPlayers")
	return 0, nil
}

func (s *stubBookingRepo) GetBookingsByMatchID(ctx context.Context, matchID uint) ([]*model.Booking, error) {
	if s.getBookingsByMatchIDFn == nil {
		return []*model.Booking{}, nil
	}
	return s.getBookingsByMatchIDFn(ctx, matchID)
}

func (s *stubBookingRepo) GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error) {
	panicUnexpectedHandlerCall("GetUserBookings")
	return nil, nil
}

func (s *stubBookingRepo) CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error) {
	panicUnexpectedHandlerCall("CancelBookingTransaction")
	return 0, nil, nil
}

func (s *stubBookingRepo) Transaction(ctx context.Context, fn func(txRepo repository.BookingRepository) error) error {
	return fn(s)
}

type stubTeamRepo struct{}

var _ repository.TeamRepository = (*stubTeamRepo)(nil)

func (s *stubTeamRepo) CreateTeam(ctx context.Context, team *model.Team) error {
	panicUnexpectedHandlerCall("CreateTeam")
	return nil
}

func (s *stubTeamRepo) GetTeamByID(ctx context.Context, teamID uint) (*model.Team, error) {
	panicUnexpectedHandlerCall("GetTeamByID")
	return nil, nil
}

func (s *stubTeamRepo) IsTeamAdmin(ctx context.Context, teamID uint, userID uint) (bool, error) {
	panicUnexpectedHandlerCall("IsTeamAdmin")
	return false, nil
}

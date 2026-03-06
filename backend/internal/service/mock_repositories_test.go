package service

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
)

func panicUnexpectedCall(name string) {
	panic("unexpected call: " + name)
}

type mockUserRepo struct {
	createUserFn        func(ctx context.Context, user *model.User) error
	getUserByIDFn       func(ctx context.Context, id uint) (*model.User, error)
	getUserByEmailFn    func(ctx context.Context, email string) (*model.User, error)
	getUserByPhoneFn    func(ctx context.Context, phone string) (*model.User, error)
	getUserByGoogleIDFn func(ctx context.Context, googleID string) (*model.User, error)
	updateUserProfileFn func(ctx context.Context, userID uint, nickname *string, avatar *string) (*model.User, error)
	updateEmailVerifyFn func(ctx context.Context, userID uint, verified bool) error
	updatePhoneVerifyFn func(ctx context.Context, userID uint, phone string, verified bool) error
}

var _ repository.UserRepository = (*mockUserRepo)(nil)

func (m *mockUserRepo) CreateUser(ctx context.Context, user *model.User) error {
	if m.createUserFn == nil {
		panicUnexpectedCall("CreateUser")
	}
	return m.createUserFn(ctx, user)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	if m.getUserByIDFn == nil {
		panicUnexpectedCall("GetUserByID")
	}
	return m.getUserByIDFn(ctx, id)
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getUserByEmailFn == nil {
		return nil, errors.New("user not found")
	}
	return m.getUserByEmailFn(ctx, email)
}

func (m *mockUserRepo) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	if m.getUserByPhoneFn == nil {
		return nil, errors.New("user not found")
	}
	return m.getUserByPhoneFn(ctx, phone)
}

func (m *mockUserRepo) GetUserByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	if m.getUserByGoogleIDFn == nil {
		return nil, errors.New("user not found")
	}
	return m.getUserByGoogleIDFn(ctx, googleID)
}

func (m *mockUserRepo) UpdateUserProfile(ctx context.Context, userID uint, nickname *string, avatar *string) (*model.User, error) {
	if m.updateUserProfileFn == nil {
		panicUnexpectedCall("UpdateUserProfile")
	}
	return m.updateUserProfileFn(ctx, userID, nickname, avatar)
}

func (m *mockUserRepo) UpdateEmailVerified(ctx context.Context, userID uint, verified bool) error {
	if m.updateEmailVerifyFn == nil {
		panicUnexpectedCall("UpdateEmailVerified")
	}
	return m.updateEmailVerifyFn(ctx, userID, verified)
}

func (m *mockUserRepo) UpdatePhoneVerified(ctx context.Context, userID uint, phone string, verified bool) error {
	if m.updatePhoneVerifyFn == nil {
		panicUnexpectedCall("UpdatePhoneVerified")
	}
	return m.updatePhoneVerifyFn(ctx, userID, phone, verified)
}

type mockMatchRepo struct {
	createMatchFn          func(ctx context.Context, match *model.Match) error
	getMatchesFn           func(ctx context.Context, filter repository.MatchFilter, offset, limit int) ([]*model.Match, int64, error)
	getMatchWithLockFn     func(ctx context.Context, matchID uint) (*model.Match, error)
	getMatchByIDFn         func(ctx context.Context, matchID uint) (*model.Match, error)
	getCommentsByMatchIDFn func(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error)
	transactionFn          func(ctx context.Context, fn func(txRepo repository.MatchRepository) error) error
}

var _ repository.MatchRepository = (*mockMatchRepo)(nil)

func (m *mockMatchRepo) CreateMatch(ctx context.Context, match *model.Match) error {
	if m.createMatchFn == nil {
		panicUnexpectedCall("CreateMatch")
	}
	return m.createMatchFn(ctx, match)
}

func (m *mockMatchRepo) GetMatches(ctx context.Context, filter repository.MatchFilter, offset, limit int) ([]*model.Match, int64, error) {
	if m.getMatchesFn == nil {
		panicUnexpectedCall("GetMatches")
	}
	return m.getMatchesFn(ctx, filter, offset, limit)
}

func (m *mockMatchRepo) GetMatchWithLock(ctx context.Context, matchID uint) (*model.Match, error) {
	if m.getMatchWithLockFn == nil {
		panicUnexpectedCall("GetMatchWithLock")
	}
	return m.getMatchWithLockFn(ctx, matchID)
}

func (m *mockMatchRepo) GetMatchByID(ctx context.Context, matchID uint) (*model.Match, error) {
	if m.getMatchByIDFn == nil {
		panicUnexpectedCall("GetMatchByID")
	}
	return m.getMatchByIDFn(ctx, matchID)
}

func (m *mockMatchRepo) GetCommentsByMatchID(ctx context.Context, matchID uint, limit int) ([]*model.Comment, error) {
	if m.getCommentsByMatchIDFn == nil {
		panicUnexpectedCall("GetCommentsByMatchID")
	}
	return m.getCommentsByMatchIDFn(ctx, matchID, limit)
}

func (m *mockMatchRepo) Transaction(ctx context.Context, fn func(txRepo repository.MatchRepository) error) error {
	if m.transactionFn != nil {
		return m.transactionFn(ctx, fn)
	}
	return fn(m)
}

type mockBookingRepo struct {
	createBookingFn            func(ctx context.Context, booking *model.Booking) error
	hasUserBookedFn            func(ctx context.Context, matchID uint, userID uint) (bool, error)
	countConfirmedPlayersFn    func(ctx context.Context, matchID uint) (int64, error)
	countWaitingPlayersFn      func(ctx context.Context, matchID uint) (int64, error)
	getBookingsByMatchIDFn     func(ctx context.Context, matchID uint) ([]*model.Booking, error)
	getUserBookingsFn          func(ctx context.Context, userID uint) ([]*model.Booking, error)
	cancelBookingTransactionFn func(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error)
	transactionFn              func(ctx context.Context, fn func(txRepo repository.BookingRepository) error) error
}

var _ repository.BookingRepository = (*mockBookingRepo)(nil)

func (m *mockBookingRepo) CreateBooking(ctx context.Context, booking *model.Booking) error {
	if m.createBookingFn == nil {
		panicUnexpectedCall("CreateBooking")
	}
	return m.createBookingFn(ctx, booking)
}

func (m *mockBookingRepo) HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error) {
	if m.hasUserBookedFn == nil {
		panicUnexpectedCall("HasUserBooked")
	}
	return m.hasUserBookedFn(ctx, matchID, userID)
}

func (m *mockBookingRepo) CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error) {
	if m.countConfirmedPlayersFn == nil {
		panicUnexpectedCall("CountConfirmedPlayers")
	}
	return m.countConfirmedPlayersFn(ctx, matchID)
}

func (m *mockBookingRepo) CountWaitingPlayers(ctx context.Context, matchID uint) (int64, error) {
	if m.countWaitingPlayersFn == nil {
		panicUnexpectedCall("CountWaitingPlayers")
	}
	return m.countWaitingPlayersFn(ctx, matchID)
}

func (m *mockBookingRepo) GetBookingsByMatchID(ctx context.Context, matchID uint) ([]*model.Booking, error) {
	if m.getBookingsByMatchIDFn == nil {
		panicUnexpectedCall("GetBookingsByMatchID")
	}
	return m.getBookingsByMatchIDFn(ctx, matchID)
}

func (m *mockBookingRepo) GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error) {
	if m.getUserBookingsFn == nil {
		panicUnexpectedCall("GetUserBookings")
	}
	return m.getUserBookingsFn(ctx, userID)
}

func (m *mockBookingRepo) CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error) {
	if m.cancelBookingTransactionFn == nil {
		panicUnexpectedCall("CancelBookingTransaction")
	}
	return m.cancelBookingTransactionFn(ctx, bookingID, userID)
}

func (m *mockBookingRepo) Transaction(ctx context.Context, fn func(txRepo repository.BookingRepository) error) error {
	if m.transactionFn != nil {
		return m.transactionFn(ctx, fn)
	}
	return fn(m)
}

type mockTeamRepo struct {
	createTeamFn  func(ctx context.Context, team *model.Team) error
	getTeamByIDFn func(ctx context.Context, teamID uint) (*model.Team, error)
	isTeamAdminFn func(ctx context.Context, teamID uint, userID uint) (bool, error)
}

var _ repository.TeamRepository = (*mockTeamRepo)(nil)

func (m *mockTeamRepo) CreateTeam(ctx context.Context, team *model.Team) error {
	if m.createTeamFn == nil {
		panicUnexpectedCall("CreateTeam")
	}
	return m.createTeamFn(ctx, team)
}

func (m *mockTeamRepo) GetTeamByID(ctx context.Context, teamID uint) (*model.Team, error) {
	if m.getTeamByIDFn == nil {
		panicUnexpectedCall("GetTeamByID")
	}
	return m.getTeamByIDFn(ctx, teamID)
}

func (m *mockTeamRepo) IsTeamAdmin(ctx context.Context, teamID uint, userID uint) (bool, error) {
	if m.isTeamAdminFn == nil {
		panicUnexpectedCall("IsTeamAdmin")
	}
	return m.isTeamAdminFn(ctx, teamID, userID)
}

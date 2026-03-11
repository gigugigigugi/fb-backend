package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"football-backend/common/config"
	"football-backend/common/utils"
	verificationcode "football-backend/common/verification"
	"football-backend/internal/model"
	pgrepo "football-backend/internal/repository/postgres"
	"football-backend/internal/router"
	"football-backend/internal/service"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// integrationSuite 聚合集成测试运行所需依赖。
type integrationSuite struct {
	db     *gorm.DB
	router *gin.Engine
}

// TestIntegrationWaitlistLimitAndCancelNoAutoPromote
// 覆盖两个关键业务约束：
// 1) WAITING 队列最多 10 人；
// 2) 取消 CONFIRMED 报名后仅通知，不自动转正。
func TestIntegrationWaitlistLimitAndCancelNoAutoPromote(t *testing.T) {
	suite := newIntegrationSuite(t)

	// 构造一个 max_players=1 的比赛，便于快速制造候补场景。
	captain := suite.mustCreateUser(t, "captain_waitlist@example.com")
	team := suite.mustCreateTeam(t, captain.ID, "Waitlist Team")
	match := suite.mustCreateMatch(t, team.ID, suite.mustCreateVenue(t, "Waitlist Field", "Tokyo", "Koto").ID, 1)

	joinUsers := make([]model.User, 0, 12)
	for i := 1; i <= 12; i++ {
		u := suite.mustCreateUser(t, fmt.Sprintf("joiner_%02d@example.com", i))
		joinUsers = append(joinUsers, u)
	}

	for i, u := range joinUsers {
		token := suite.mustToken(t, u.ID)
		resp := suite.doJSONRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/matches/%d/join", match.ID), "", token)
		// 前 11 人应成功（1 CONFIRMED + 10 WAITING）。
		if i < 11 && resp.Code != http.StatusOK {
			t.Fatalf("expected join success for index=%d, got status=%d body=%s", i, resp.Code, resp.Body.String())
		}
		// 第 12 人应触发 WAITING 已满。
		if i == 11 && resp.Code != http.StatusBadRequest {
			t.Fatalf("expected waitlist full for last user, got status=%d body=%s", resp.Code, resp.Body.String())
		}
	}

	var confirmedCount int64
	var waitingCount int64
	if err := suite.db.Model(&model.Booking{}).Where("match_id = ? AND status = ?", match.ID, "CONFIRMED").Count(&confirmedCount).Error; err != nil {
		t.Fatalf("count confirmed failed: %v", err)
	}
	if err := suite.db.Model(&model.Booking{}).Where("match_id = ? AND status = ?", match.ID, "WAITING").Count(&waitingCount).Error; err != nil {
		t.Fatalf("count waiting failed: %v", err)
	}
	if confirmedCount != 1 || waitingCount != 10 {
		t.Fatalf("unexpected roster counts confirmed=%d waiting=%d", confirmedCount, waitingCount)
	}

	var captainBooking model.Booking
	if err := suite.db.Where("match_id = ? AND user_id = ?", match.ID, joinUsers[0].ID).First(&captainBooking).Error; err != nil {
		t.Fatalf("find captain booking failed: %v", err)
	}
	cancelToken := suite.mustToken(t, joinUsers[0].ID)
	cancelResp := suite.doJSONRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/cancel", captainBooking.ID), "", cancelToken)
	if cancelResp.Code != http.StatusOK {
		t.Fatalf("expected cancel success, got status=%d body=%s", cancelResp.Code, cancelResp.Body.String())
	}

	// 核心断言：取消后不自动转正，WAITING 人数保持不变。
	if err := suite.db.Model(&model.Booking{}).Where("match_id = ? AND status = ?", match.ID, "CONFIRMED").Count(&confirmedCount).Error; err != nil {
		t.Fatalf("count confirmed after cancel failed: %v", err)
	}
	if err := suite.db.Model(&model.Booking{}).Where("match_id = ? AND status = ?", match.ID, "WAITING").Count(&waitingCount).Error; err != nil {
		t.Fatalf("count waiting after cancel failed: %v", err)
	}
	if confirmedCount != 0 || waitingCount != 10 {
		t.Fatalf("expected no auto promote after cancel, confirmed=%d waiting=%d", confirmedCount, waitingCount)
	}
}

// TestIntegrationVenuesEndpoints
// 验证 venues 发现接口在真实数据库上的输出结构与过滤行为。
func TestIntegrationVenuesEndpoints(t *testing.T) {
	suite := newIntegrationSuite(t)

	// 准备跨行政区测试数据。
	suite.mustCreateVenue(t, "Shibuya A", "Tokyo", "Shibuya")
	suite.mustCreateVenue(t, "Shibuya B", "Tokyo", "Shibuya")
	suite.mustCreateVenue(t, "Koto A", "Tokyo", "Koto")
	suite.mustCreateVenue(t, "Yokohama A", "Kanagawa", "Yokohama")

	regionsResp := suite.doJSONRequest(t, http.MethodGet, "/api/v1/venues/regions", "", "")
	if regionsResp.Code != http.StatusOK {
		t.Fatalf("expected regions 200, got status=%d body=%s", regionsResp.Code, regionsResp.Body.String())
	}
	var regionsBody struct {
		Code int `json:"code"`
		Data []struct {
			Prefecture string `json:"prefecture"`
			VenueCount int64  `json:"venue_count"`
			Cities     []struct {
				City       string `json:"city"`
				VenueCount int64  `json:"venue_count"`
			} `json:"cities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(regionsResp.Body.Bytes(), &regionsBody); err != nil {
		t.Fatalf("unmarshal regions response failed: %v", err)
	}
	if regionsBody.Code != 0 || len(regionsBody.Data) == 0 {
		t.Fatalf("unexpected regions response: %+v", regionsBody)
	}

	// 验证 map 接口的 prefecture/city 过滤是否生效。
	mapResp := suite.doJSONRequest(t, http.MethodGet, "/api/v1/venues/map?prefecture=Tokyo&city=Shibuya&limit=10", "", "")
	if mapResp.Code != http.StatusOK {
		t.Fatalf("expected map 200, got status=%d body=%s", mapResp.Code, mapResp.Body.String())
	}
	var mapBody struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				ID         uint   `json:"id"`
				Name       string `json:"name"`
				Prefecture string `json:"prefecture"`
				City       string `json:"city"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(mapResp.Body.Bytes(), &mapBody); err != nil {
		t.Fatalf("unmarshal map response failed: %v", err)
	}
	if mapBody.Code != 0 || mapBody.Data.Total != 2 {
		t.Fatalf("expected Tokyo/Shibuya total=2, got %+v", mapBody.Data)
	}
}

// TestIntegrationSettlementAndSubteamsPermission
// 覆盖赛后接口在不同角色下的权限边界与落库结果。
func TestIntegrationSettlementAndSubteamsPermission(t *testing.T) {
	suite := newIntegrationSuite(t)

	captain := suite.mustCreateUser(t, "captain_settle@example.com")
	admin := suite.mustCreateUser(t, "admin_settle@example.com")
	member := suite.mustCreateUser(t, "member_settle@example.com")
	team := suite.mustCreateTeam(t, captain.ID, "Settle Team")
	suite.mustCreateTeamMember(t, team.ID, admin.ID, "ADMIN")
	suite.mustCreateTeamMember(t, team.ID, member.ID, "MEMBER")

	match := suite.mustCreateMatch(t, team.ID, suite.mustCreateVenue(t, "Settle Field", "Tokyo", "Shinjuku").ID, 14)
	b1 := suite.mustCreateBooking(t, match.ID, captain.ID, "CONFIRMED")
	b2 := suite.mustCreateBooking(t, match.ID, admin.ID, "CONFIRMED")

	// 队长可执行 settlement。
	captainToken := suite.mustToken(t, captain.ID)
	settleBody := `{"payment_status":"PAID","booking_ids":[` + fmt.Sprintf("%d,%d", b1.ID, b2.ID) + `]}`
	settleResp := suite.doJSONRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/matches/%d/settlement", match.ID), settleBody, captainToken)
	if settleResp.Code != http.StatusOK {
		t.Fatalf("expected settlement 200, got status=%d body=%s", settleResp.Code, settleResp.Body.String())
	}

	var paidCount int64
	if err := suite.db.Model(&model.Booking{}).
		Where("match_id = ? AND payment_status = ?", match.ID, "PAID").
		Count(&paidCount).Error; err != nil {
		t.Fatalf("count PAID bookings failed: %v", err)
	}
	if paidCount != 2 {
		t.Fatalf("expected 2 PAID bookings, got %d", paidCount)
	}

	// MEMBER 无权限分队，应返回 403。
	memberToken := suite.mustToken(t, member.ID)
	forbiddenResp := suite.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/api/v1/matches/%d/subteams", match.ID),
		fmt.Sprintf(`{"assignments":[{"booking_id":%d,"sub_team":"A"}]}`, b1.ID),
		memberToken,
	)
	if forbiddenResp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for member subteams, got status=%d body=%s", forbiddenResp.Code, forbiddenResp.Body.String())
	}

	// 不带 token 应命中认证失败，返回 401。
	unauthorizedResp := suite.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/api/v1/matches/%d/subteams", match.ID),
		fmt.Sprintf(`{"assignments":[{"booking_id":%d,"sub_team":"A"}]}`, b1.ID),
		"",
	)
	if unauthorizedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got status=%d body=%s", unauthorizedResp.Code, unauthorizedResp.Body.String())
	}

	// ADMIN 可执行分队，且结果应正确落库。
	adminToken := suite.mustToken(t, admin.ID)
	successResp := suite.doJSONRequest(
		t,
		http.MethodPost,
		fmt.Sprintf("/api/v1/matches/%d/subteams", match.ID),
		fmt.Sprintf(`{"assignments":[{"booking_id":%d,"sub_team":"A"},{"booking_id":%d,"sub_team":"B"}]}`, b1.ID, b2.ID),
		adminToken,
	)
	if successResp.Code != http.StatusOK {
		t.Fatalf("expected admin subteams 200, got status=%d body=%s", successResp.Code, successResp.Body.String())
	}

	var check1, check2 model.Booking
	if err := suite.db.First(&check1, b1.ID).Error; err != nil {
		t.Fatalf("load booking1 failed: %v", err)
	}
	if err := suite.db.First(&check2, b2.ID).Error; err != nil {
		t.Fatalf("load booking2 failed: %v", err)
	}
	if check1.SubTeam != "A" || check2.SubTeam != "B" {
		t.Fatalf("unexpected subteam result: b1=%s b2=%s", check1.SubTeam, check2.SubTeam)
	}
}

// newIntegrationSuite 创建每个测试用例的隔离运行环境。
// 如果未配置 INTEGRATION_TEST_DSN，则自动跳过，避免影响默认单测流程。
func newIntegrationSuite(t *testing.T) *integrationSuite {
	t.Helper()

	dsn := resolveIntegrationDSN()
	if dsn == "" {
		t.Skip("skip integration test: INTEGRATION_TEST_DSN/DB_DSN is empty")
	}
	if !isSafeIntegrationDSN(dsn) {
		t.Skip("skip integration test: resolved DB is not a test database; set INTEGRATION_TEST_DSN to *_test or set INTEGRATION_ALLOW_NON_TEST_DB=true")
	}
	if err := ensureDatabaseExists(dsn); err != nil {
		t.Fatalf("prepare integration database failed: %v", err)
	}

	// 集成测试使用真实 PostgreSQL，日志设为 Silent 避免噪音。
	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("connect postgres failed: %v", err)
	}

	resetDatabase(t, db)
	r := setupRouterForIntegration(t, db)
	return &integrationSuite{
		db:     db,
		router: r,
	}
}

// resolveIntegrationDSN 解析集成测试数据库连接串。
// 优先级：
// 1) INTEGRATION_TEST_DSN（环境变量）
// 2) .env 中的 INTEGRATION_TEST_DSN
// 3) .env 中的 DB_DSN（作为回退）
func resolveIntegrationDSN() string {
	dsn := cleanDSN(os.Getenv("INTEGRATION_TEST_DSN"))
	if dsn != "" {
		return dsn
	}

	// go test 执行时工作目录可能在 test/integration，需要向上回溯到 backend/.env。
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	dsn = cleanDSN(os.Getenv("INTEGRATION_TEST_DSN"))
	if dsn != "" {
		return dsn
	}
	return cleanDSN(os.Getenv("DB_DSN"))
}

// isSafeIntegrationDSN 防止集成测试误删非测试库数据。
// 默认要求库名包含 "test"；若明确允许，可设置 INTEGRATION_ALLOW_NON_TEST_DB=true 覆盖该保护。
func isSafeIntegrationDSN(dsn string) bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("INTEGRATION_ALLOW_NON_TEST_DB")), "true") {
		return true
	}

	dbName := strings.ToLower(strings.TrimSpace(extractDBName(dsn)))
	return dbName != "" && strings.Contains(dbName, "test")
}

// extractDBName 从 PostgreSQL DSN 中提取库名。
// 支持：
// 1) key/value 形式：host=... dbname=football_test ...
// 2) URL 形式：postgres://user:pass@host:5432/football_test?sslmode=disable
func extractDBName(dsn string) string {
	trimmed := cleanDSN(dsn)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "postgres://") || strings.HasPrefix(trimmed, "postgresql://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return ""
		}
		return strings.TrimPrefix(u.Path, "/")
	}

	for _, part := range strings.Fields(trimmed) {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) != 2 {
			continue
		}
		if strings.EqualFold(pair[0], "dbname") {
			return strings.TrimSpace(pair[1])
		}
	}
	return ""
}

// cleanDSN 清理 DSN 文本中的空格与包裹引号。
func cleanDSN(raw string) string {
	dsn := strings.TrimSpace(raw)
	dsn = strings.Trim(dsn, `"`)
	return strings.TrimSpace(dsn)
}

// ensureDatabaseExists 确保目标数据库存在。
// 若测试库不存在，会尝试用同一账号连接 postgres 库并执行 CREATE DATABASE。
func ensureDatabaseExists(targetDSN string) error {
	targetDB := strings.TrimSpace(extractDBName(targetDSN))
	if targetDB == "" {
		return errors.New("cannot parse dbname from DSN")
	}

	adminDSN, err := replaceDSNDatabase(targetDSN, "postgres")
	if err != nil {
		return err
	}

	adminDB, err := gorm.Open(pgdriver.Open(adminDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("connect admin database failed: %w", err)
	}

	var exists int64
	if err := adminDB.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", targetDB).Scan(&exists).Error; err != nil {
		return fmt.Errorf("check database exists failed: %w", err)
	}
	if exists > 0 {
		return nil
	}

	createSQL := fmt.Sprintf(`CREATE DATABASE "%s"`, escapeIdentifier(targetDB))
	if err := adminDB.Exec(createSQL).Error; err != nil {
		return fmt.Errorf("create database %q failed: %w", targetDB, err)
	}
	return nil
}

// replaceDSNDatabase 将 DSN 中的数据库名替换为 newDBName。
// 支持 URL 形式与 key/value 形式。
func replaceDSNDatabase(dsn, newDBName string) (string, error) {
	trimmed := cleanDSN(dsn)
	if trimmed == "" {
		return "", errors.New("empty DSN")
	}

	if strings.HasPrefix(trimmed, "postgres://") || strings.HasPrefix(trimmed, "postgresql://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("parse postgres URL DSN failed: %w", err)
		}
		u.Path = "/" + newDBName
		return u.String(), nil
	}

	parts := strings.Fields(trimmed)
	found := false
	for i, part := range parts {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) != 2 {
			continue
		}
		if strings.EqualFold(pair[0], "dbname") {
			parts[i] = pair[0] + "=" + newDBName
			found = true
			break
		}
	}
	if !found {
		parts = append(parts, "dbname="+newDBName)
	}
	return strings.Join(parts, " "), nil
}

// escapeIdentifier 对 SQL 标识符做最小转义，防止引号破坏 SQL。
func escapeIdentifier(name string) string {
	return strings.ReplaceAll(name, `"`, `""`)
}

// setupRouterForIntegration 组装与生产分层一致的路由与依赖。
// 这里显式关闭 AuthBypass，用于验证 401/403 权限行为。
func setupRouterForIntegration(t *testing.T, db *gorm.DB) *gin.Engine {
	t.Helper()

	// 统一集成测试配置：开启真实 JWT 鉴权，便于覆盖 401/403 边界。
	config.App = &config.AppConfig{
		Env:        "test",
		Port:       "0",
		AuthBypass: false,
		DB:         config.DBConfig{DSN: ""},
		JWT: config.JWTConfig{
			Secret: "integration-secret",
			Exp:    72,
		},
		Verification: config.VerificationConfig{
			Provider:       "mock",
			HTTPTimeoutSec: 2,
		},
	}

	userRepo := pgrepo.NewUserRepository(db)
	teamRepo := pgrepo.NewTeamRepository(db)
	matchRepo := pgrepo.NewMatchRepository(db)
	bookingRepo := pgrepo.NewBookingRepository(db)
	venueRepo := pgrepo.NewVenueRepository(db)
	verificationRepo := pgrepo.NewVerificationRepository(db)

	matchSvc := service.NewMatchService(matchRepo, bookingRepo, teamRepo, userRepo, nil)
	teamSvc := service.NewTeamService(teamRepo)
	authSvc := service.NewAuthService(userRepo, verificationRepo, verificationcode.NewMockCodeProvider())
	userSvc := service.NewUserService(userRepo)
	venueSvc := service.NewVenueService(venueRepo)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.SetupRouter(r, matchSvc, teamSvc, authSvc, userSvc, venueSvc)
	return r
}

// resetDatabase 在每个用例开始前重建核心表，保证测试可重复且相互隔离。
func resetDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Migrator().DropTable(
		&model.Match{},
		&model.TeamMember{},
		&model.Team{},
		&model.User{},
		&model.Booking{},
		&model.Venue{},
		&model.Comment{},
		&model.VerificationChallenge{},
	); err != nil {
		t.Fatalf("drop tables failed: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.Venue{},
		&model.Match{},
		&model.Booking{},
		&model.Comment{},
		&model.VerificationChallenge{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
}

// doJSONRequest 使用 httptest 对完整 Gin 路由发起请求，模拟真实 API 调用。
func (s *integrationSuite) doJSONRequest(t *testing.T, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()

	var payloadReader *bytes.Reader
	if body == "" {
		payloadReader = bytes.NewReader(nil)
	} else {
		payloadReader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, payloadReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// mustToken 生成测试用户 JWT；失败即中断当前测试。
func (s *integrationSuite) mustToken(t *testing.T, userID uint) string {
	t.Helper()
	token, err := utils.GenerateToken(userID)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}
	return token
}

// mustCreateUser 创建基础测试用户。
func (s *integrationSuite) mustCreateUser(t *testing.T, email string) model.User {
	t.Helper()
	user := model.User{
		Email:        email,
		PasswordHash: "integration_hash",
		Nickname:     "integration-user",
		Reputation:   100,
	}
	if err := s.db.WithContext(context.Background()).Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	return user
}

// mustCreateTeam 创建测试球队（captain_id 为创建者）。
func (s *integrationSuite) mustCreateTeam(t *testing.T, captainID uint, name string) model.Team {
	t.Helper()
	team := model.Team{
		Name:      name,
		CaptainID: captainID,
	}
	if err := s.db.WithContext(context.Background()).Create(&team).Error; err != nil {
		t.Fatalf("create team failed: %v", err)
	}
	return team
}

// mustCreateTeamMember 创建球队成员关系，用于构造 OWNER/ADMIN/MEMBER 权限场景。
func (s *integrationSuite) mustCreateTeamMember(t *testing.T, teamID, userID uint, role string) {
	t.Helper()
	member := model.TeamMember{
		TeamID:   teamID,
		UserID:   userID,
		Role:     role,
		JoinTime: time.Now().UTC(),
	}
	if err := s.db.WithContext(context.Background()).Create(&member).Error; err != nil {
		t.Fatalf("create team member failed: %v", err)
	}
}

// mustCreateVenue 创建具备经纬度的测试场地，便于 /venues/map 查询。
func (s *integrationSuite) mustCreateVenue(t *testing.T, name, prefecture, city string) model.Venue {
	t.Helper()
	venue := model.Venue{
		Name:       name,
		Prefecture: prefecture,
		City:       city,
		Address:    name + " address",
		Latitude:   35.6,
		Longitude:  139.7,
		IsVerified: true,
	}
	if err := s.db.WithContext(context.Background()).Create(&venue).Error; err != nil {
		t.Fatalf("create venue failed: %v", err)
	}
	return venue
}

// mustCreateMatch 创建 RECRUITING 状态测试比赛。
func (s *integrationSuite) mustCreateMatch(t *testing.T, teamID, venueID uint, maxPlayers int) model.Match {
	t.Helper()
	startAt := time.Now().UTC().Add(48 * time.Hour)
	match := model.Match{
		TeamID:     teamID,
		VenueID:    venueID,
		StartTime:  startAt,
		EndTime:    startAt.Add(2 * time.Hour),
		Price:      1000,
		MaxPlayers: maxPlayers,
		Format:     7,
		Note:       "integration match",
		Status:     "RECRUITING",
	}
	if err := s.db.WithContext(context.Background()).Create(&match).Error; err != nil {
		t.Fatalf("create match failed: %v", err)
	}
	return match
}

// mustCreateBooking 创建测试报名记录（常用于赛后 settlement/subteams 用例）。
func (s *integrationSuite) mustCreateBooking(t *testing.T, matchID, userID uint, status string) model.Booking {
	t.Helper()
	booking := model.Booking{
		MatchID:       matchID,
		UserID:        userID,
		Status:        status,
		PaymentStatus: "UNPAID",
	}
	if err := s.db.WithContext(context.Background()).Create(&booking).Error; err != nil {
		t.Fatalf("create booking failed: %v", err)
	}
	return booking
}

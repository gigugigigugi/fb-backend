package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"football-backend/internal/model"

	"github.com/joho/godotenv"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var targetTables = []string{
	"users",
	"venues",
	"teams",
	"team_members",
	"matches",
	"bookings",
	"comments",
	"verification_challenges",
}

func main() {
	dsn := resolveSchemaCheckDSN()
	if dsn == "" {
		fail("SCHEMA_CHECK_DSN/INTEGRATION_TEST_DSN/DB_DSN is empty")
	}

	adminDSN, err := replaceDSNDatabase(dsn, "postgres")
	if err != nil {
		fail("build admin DSN failed: %v", err)
	}

	adminDB, err := openGorm(adminDSN)
	if err != nil {
		fail("connect admin database failed: %v", err)
	}
	defer closeGorm(adminDB)

	suffix := randomSuffix()
	initDBName := "schema_init_" + suffix
	modelDBName := "schema_model_" + suffix

	if err := createDatabase(adminDB, initDBName); err != nil {
		fail("create init database failed: %v", err)
	}
	if err := createDatabase(adminDB, modelDBName); err != nil {
		_ = dropDatabase(adminDB, initDBName)
		fail("create model database failed: %v", err)
	}

	defer func() {
		_ = dropDatabase(adminDB, initDBName)
		_ = dropDatabase(adminDB, modelDBName)
	}()

	initDBDSN, err := replaceDSNDatabase(dsn, initDBName)
	if err != nil {
		fail("build init DB DSN failed: %v", err)
	}
	modelDBDSN, err := replaceDSNDatabase(dsn, modelDBName)
	if err != nil {
		fail("build model DB DSN failed: %v", err)
	}

	initDB, err := openGorm(initDBDSN)
	if err != nil {
		fail("connect init database failed: %v", err)
	}
	if err := applyInitSQL(initDB, findInitSQLPath()); err != nil {
		closeGorm(initDB)
		fail("apply init SQL failed: %v", err)
	}
	initSnapshot, err := loadSchemaSnapshot(initDB, targetTables)
	closeGorm(initDB)
	if err != nil {
		fail("load init snapshot failed: %v", err)
	}

	modelDB, err := openGorm(modelDBDSN)
	if err != nil {
		fail("connect model database failed: %v", err)
	}
	if err := applyAutoMigrate(modelDB); err != nil {
		closeGorm(modelDB)
		fail("apply automigrate failed: %v", err)
	}
	modelSnapshot, err := loadSchemaSnapshot(modelDB, targetTables)
	closeGorm(modelDB)
	if err != nil {
		fail("load model snapshot failed: %v", err)
	}

	diffs := diffSnapshots(initSnapshot, modelSnapshot, targetTables)
	if len(diffs) > 0 {
		fmt.Println("SCHEMA CHECK FAILED")
		for _, d := range diffs {
			fmt.Println("- " + d)
		}
		os.Exit(1)
	}

	fmt.Println("SCHEMA CHECK PASSED: init.sql and AutoMigrate are aligned on target tables")
}

func resolveSchemaCheckDSN() string {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	candidates := []string{
		os.Getenv("SCHEMA_CHECK_DSN"),
		os.Getenv("INTEGRATION_TEST_DSN"),
		os.Getenv("DB_DSN"),
	}
	for _, raw := range candidates {
		if dsn := cleanDSN(raw); dsn != "" {
			return dsn
		}
	}
	return ""
}

func openGorm(dsn string) (*gorm.DB, error) {
	return gorm.Open(pgdriver.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
}

func applyInitSQL(db *gorm.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s failed: %w", path, err)
	}
	return db.Exec(string(content)).Error
}

func applyAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.Venue{},
		&model.Match{},
		&model.Booking{},
		&model.Comment{},
		&model.VerificationChallenge{},
	)
}

func findInitSQLPath() string {
	paths := []string{
		filepath.FromSlash("sql/init.sql"),
		filepath.FromSlash("../backend/sql/init.sql"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.FromSlash("sql/init.sql")
}

func loadSchemaSnapshot(db *gorm.DB, tables []string) (map[string]map[string]struct{}, error) {
	snapshot := make(map[string]map[string]struct{}, len(tables))
	for _, table := range tables {
		exists, err := tableExists(db, table)
		if err != nil {
			return nil, err
		}
		if !exists {
			snapshot[table] = map[string]struct{}{}
			continue
		}

		cols, err := loadColumns(db, table)
		if err != nil {
			return nil, err
		}
		snapshot[table] = cols
	}
	return snapshot, nil
}

func tableExists(db *gorm.DB, table string) (bool, error) {
	var exists bool
	err := db.Raw(`
SELECT EXISTS (
  SELECT 1
  FROM information_schema.tables
  WHERE table_schema = 'public' AND table_name = ?
)`, table).Scan(&exists).Error
	return exists, err
}

func loadColumns(db *gorm.DB, table string) (map[string]struct{}, error) {
	type row struct {
		ColumnName string `gorm:"column:column_name"`
	}
	rows := make([]row, 0)
	err := db.Raw(`
SELECT column_name
FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = ?
ORDER BY column_name`, table).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	cols := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		cols[strings.ToLower(strings.TrimSpace(r.ColumnName))] = struct{}{}
	}
	return cols, nil
}

func diffSnapshots(initSnap, modelSnap map[string]map[string]struct{}, tables []string) []string {
	diffs := make([]string, 0)
	for _, table := range tables {
		initCols := initSnap[table]
		modelCols := modelSnap[table]

		if len(initCols) == 0 {
			diffs = append(diffs, fmt.Sprintf("table %s missing in init.sql snapshot", table))
			continue
		}
		if len(modelCols) == 0 {
			diffs = append(diffs, fmt.Sprintf("table %s missing in AutoMigrate snapshot", table))
			continue
		}

		missingInInit := diffColumnSet(modelCols, initCols)
		missingInModel := diffColumnSet(initCols, modelCols)
		if len(missingInInit) > 0 {
			diffs = append(diffs, fmt.Sprintf("table %s columns missing in init.sql: %s", table, strings.Join(missingInInit, ", ")))
		}
		if len(missingInModel) > 0 {
			diffs = append(diffs, fmt.Sprintf("table %s columns missing in AutoMigrate: %s", table, strings.Join(missingInModel, ", ")))
		}
	}
	return diffs
}

func diffColumnSet(source, target map[string]struct{}) []string {
	out := make([]string, 0)
	for c := range source {
		if _, ok := target[c]; !ok {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

func createDatabase(adminDB *gorm.DB, dbName string) error {
	sql := fmt.Sprintf(`CREATE DATABASE "%s"`, escapeIdentifier(dbName))
	return adminDB.Exec(sql).Error
}

func dropDatabase(adminDB *gorm.DB, dbName string) error {
	if err := adminDB.Exec(`
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = ? AND pid <> pg_backend_pid()`, dbName).Error; err != nil {
		return err
	}
	sql := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, escapeIdentifier(dbName))
	return adminDB.Exec(sql).Error
}

func closeGorm(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

func replaceDSNDatabase(dsn, newDBName string) (string, error) {
	trimmed := cleanDSN(dsn)
	if trimmed == "" {
		return "", fmt.Errorf("empty DSN")
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

func cleanDSN(raw string) string {
	dsn := strings.TrimSpace(raw)
	dsn = strings.Trim(dsn, `"`)
	return strings.TrimSpace(dsn)
}

func escapeIdentifier(name string) string {
	return strings.ReplaceAll(name, `"`, `""`)
}

func randomSuffix() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%d_%d", time.Now().Unix(), r.Intn(100000))
}

func fail(format string, args ...any) {
	fmt.Printf("SCHEMA CHECK ERROR: "+format+"\n", args...)
	os.Exit(1)
}

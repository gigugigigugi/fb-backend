# Integration Tests

该目录用于运行基于真实 PostgreSQL 的集成测试，覆盖核心业务链路与权限边界。

## 1. 环境变量

推荐在运行前设置：

```bash
export INTEGRATION_TEST_DSN='host=127.0.0.1 user=postgres password=postgres dbname=football_test port=5432 sslmode=disable TimeZone=Asia/Tokyo'
```

PowerShell 示例：

```powershell
$env:INTEGRATION_TEST_DSN = "host=127.0.0.1 user=postgres password=postgres dbname=football_test port=5432 sslmode=disable TimeZone=Asia/Tokyo"
```

## 2. 执行命令

```bash
go test ./test/integration -count=1 -v
```

如果未设置 `INTEGRATION_TEST_DSN`，测试会自动尝试读取 `.env` 中的 `INTEGRATION_TEST_DSN`，再回退到 `DB_DSN`。  
若仍为空则自动 `skip`，不会影响默认 `go test ./...`。

## 3. 安全保护

测试会执行 `DropTable + AutoMigrate`，默认只允许库名包含 `test` 的数据库。  
若解析到非测试库（例如 `football`），会自动 `skip`，防止误删数据。

若目标测试库不存在，测试会尝试自动创建（通过同账号连接 `postgres` 库执行 `CREATE DATABASE`）。  
如果账号无建库权限，会在启动阶段报错，请手动建库后重试。

如确需在非 test 库执行（不推荐），可显式设置：

```powershell
$env:INTEGRATION_ALLOW_NON_TEST_DB = "true"
```

## 4. 当前覆盖

1. WAITING 上限（10）与取消后不自动转正。  
2. `GET /api/v1/venues/regions` 与 `GET /api/v1/venues/map`。  
3. `POST /matches/:id/settlement`、`POST /matches/:id/subteams` 的 200/403/401 权限边界。

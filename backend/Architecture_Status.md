# Backend Architecture & Status

## 1. 架构概览

当前后端采用标准分层：
- `handler`：HTTP 参数解析与响应封装
- `service`：业务编排与规则校验
- `repository`：数据访问抽象
- `repository/postgres`：PostgreSQL 实现
- `model`：Gorm 实体

请求流：
`Client -> Router -> Handler -> Service -> Repository -> PostgreSQL`

## 2. 当前已落地能力

1. 统一 API 前缀：`/api/v1`。
2. 鉴权：Bearer Token + 开发环境可选 `AUTH_BYPASS`。
3. 比赛主流程：列表、报名、批量建赛、取消报名、详情聚合。
4. 用户主流程：注册、登录、Google 登录、我的资料 GET/PUT。
5. 验证码：
   - 发送频控（冷却+窗口）
   - 连续失败锁定
   - 状态持久化到 `verification_challenges`
   - 发送通道支持 `mock/real` provider 切换
6. 通知系统：异步 Dispatcher，单次投递，无自动重试。

## 3. 关键业务约束

1. WAITING 队列上限：10。
2. 取消报名后不自动转正，仅通知 WAITING 用户。
3. 候补通知最多处理前 10 位用户（FIFO）。
4. 通知仅使用已验证联系方式（email_verified / phone_verified）。

## 4. 测试现状

1. `service` 层已有关键单测：
   - users/me 更新边界
   - matches/:id/details 的 user_status 与聚合
   - CancelBooking 最多通知 10 人
2. `handler` 层测试已补核心新 API（成功/失败、状态码、响应结构）。
3. 集成测试已落地并可执行：
   - WAITING 上限与取消后不自动转正
   - venues regions/map
   - settlement/subteams 权限边界
4. 已新增 `cmd/schema-check`：
   - 对比 `sql/init.sql` 与 `AutoMigrate` 产物的核心表字段集合
   - 不一致时直接失败
5. GitHub Actions 已接入三阶段：
   - Unit Tests
   - Integration Tests
   - Schema Check

## 5. 待完成项（P2）

1. 真实通知 provider 的联调与验收（real 模式）：
   - 已提供独立联调命令：`go run ./cmd/provider-smoke`
   - 已提供 GitHub 手动联调工作流：`Provider Smoke`
   - 待完成：使用供应商测试账号完成邮件/短信双通道实测并留痕
2. 发布脚本收敛（部署前检查、回滚策略、环境一致性）。
3. 前端落地与端到端联调。

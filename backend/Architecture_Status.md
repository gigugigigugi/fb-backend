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
3. 当前 `go test ./...` 可通过。

## 5. 待完成项（P2）

1. 真实通知 provider 的联调与验收（real 模式）。
2. 集成测试与发布脚本收敛。
3. 前端落地与端到端联调。

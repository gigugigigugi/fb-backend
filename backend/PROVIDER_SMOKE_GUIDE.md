# Verification Provider 联调验收指南（P2-3）

本文档用于在 `VERIFICATION_PROVIDER=real` 模式下，独立验证邮件/短信网关连通性与凭证配置是否正确。

## 1. 准备配置

在 `.env` 中至少配置以下变量：

```env
VERIFICATION_PROVIDER=real
VERIFICATION_HTTP_TIMEOUT_SEC=8

VERIFICATION_EMAIL_API_URL=https://your-email-provider/send
VERIFICATION_EMAIL_API_KEY=your-email-api-key
VERIFICATION_EMAIL_FROM=no-reply@example.com

VERIFICATION_SMS_API_URL=https://your-sms-provider/send
VERIFICATION_SMS_API_KEY=your-sms-api-key
VERIFICATION_SMS_SENDER=YourBrand
```

## 2. 配置联调目标

新增（或临时导出）以下变量：

```env
PROVIDER_SMOKE_CHANNELS=both
PROVIDER_SMOKE_EMAIL_TO=your-test-email@example.com
PROVIDER_SMOKE_PHONE_TO=+819012345678
PROVIDER_SMOKE_CODE=654321
```

说明：
- `PROVIDER_SMOKE_CHANNELS`：`email` / `sms` / `both`（默认 `both`）。
- `PROVIDER_SMOKE_CODE`：可省略，默认 `654321`。

## 3. 运行联调命令

在 `backend` 目录执行：

```bash
go run ./cmd/provider-smoke
```

预期输出：
- 所选通道全部成功时：`provider smoke success: 所选通道均已通过`
- 任一通道失败时：进程退出码非 0，并输出具体 HTTP 错误信息（状态码/响应体）

## 4. 验收结论建议

1. 至少完成一次 `email`、一次 `sms` 的独立验证。  
2. 在测试环境保留 provider 返回日志（含 request id）。  
3. 完成后再在业务接口层验证：
   - `POST /api/v1/users/me/verify/email/send`
   - `POST /api/v1/users/me/verify/phone/send`

## 5. GitHub Actions 手动联调（可选）

仓库已提供手动工作流：`Provider Smoke`。  
你可以在 GitHub 页面通过 `Run workflow` 触发，不依赖本地机器。

需要先在仓库 Secrets 中配置：

- `VERIFICATION_EMAIL_API_URL`
- `VERIFICATION_EMAIL_API_KEY`
- `VERIFICATION_EMAIL_FROM`
- `VERIFICATION_SMS_API_URL`
- `VERIFICATION_SMS_API_KEY`
- `VERIFICATION_SMS_SENDER`
- `PROVIDER_SMOKE_EMAIL_TO`
- `PROVIDER_SMOKE_PHONE_TO`

运行时输入：
- `channels`: `email` / `sms` / `both`
- `code`: 验证码内容（用于日志定位）

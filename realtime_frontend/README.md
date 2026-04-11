# realtime_frontend_fixed

这是修复后的完整前端项目，可直接替换。

## 已修复的问题

- 静态资源 embed 路径错误，导致 css/js 404 或 MIME type 异常
- 下单字段 `auto_cashout_multiple` 拼写错误
- 页面未自动计算当前倍率
- 页面未按状态联动下注/兑现按钮

## 功能

- 当前局状态展示
- 本地实时倍率计算
- pre_start 和 flying 阶段允许下注
- flying 阶段允许兑现
- 下单
- 手动兑现
- 我的订单列表
- 操作日志

## 环境变量

- `FRONTEND_LISTEN_ADDR`：前端监听地址，默认 `:8090`
- `GAME_BACKEND_URL`：后端 API 地址，默认 `http://127.0.0.1:8888`

## 启动

```bash
go run .
```

或者：

```bash
FRONTEND_LISTEN_ADDR=:8090 GAME_BACKEND_URL=http://127.0.0.1:8888 go run .
```

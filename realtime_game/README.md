# 实时游戏服务（彻底重写版）

这个目录是一套全新的实时游戏实现。

## 设计目标

- 只保留现有 **MySQL 表结构 / Model / ApiSys 协议**。
- 不复用旧的下注、兑现、状态机主逻辑。
- 应用层只需要运行两个 Go 进程：
  - `realtime-api`
  - `realtime-worker`
- 支持多副本部署。
- Redis 负责热状态，MySQL 负责最终账。
- 前端通过 `current-round` 拉取当前局快照，本地自行计算展示倍率。

## 目录说明

- `cmd/realtime-api`：HTTP 接口服务
- `cmd/realtime-worker`：局调度与自动兑现服务
- `config`：配置结构
- `context`：依赖装配
- `domain`：领域模型与通用计算
- `settlement`：ApiSys 适配层
- `service`：业务服务
- `store`：Redis 热状态与租约
- `types`：接口请求与响应结构

## 运行方式

### 1. 启动 API

```bash
go run ./realtime_new/cmd/realtime-api -f realtime_new/etc/realtime-api.yaml
```

### 2. 启动 Worker

```bash
go run ./realtime_new/cmd/realtime-worker -f realtime_new/etc/realtime-worker.yaml
```

## 当前覆盖的能力

- 当前局创建与轮转
- 当前局快照读取
- 下单
- 手动兑现（滚盘全兑 / 赛前半兑 / 赛前全兑）
- 自动兑现
- 局结束收口
- 失败重试（兑现 / 退款）
- Redis owner lease，支持多 worker 抢占容灾

## 重要说明

1. 这套代码**不会修改现有表结构**。
2. 为了保证 DB 字段兼容，`bet`、`crash_term` 等表的填充规则延续旧系统的缩放约定：
   - 金额字段：`*10000`
   - 倍数字段：`*100*10000`
   - 当前局倍数：`*100`
3. 这套代码默认关闭 bonus / jackpot，后续如果需要，可以在 `close_round_service.go` 里继续扩展。

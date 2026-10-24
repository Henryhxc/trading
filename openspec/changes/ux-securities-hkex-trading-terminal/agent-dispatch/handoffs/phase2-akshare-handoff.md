# Phase 2 AkShare / Deployment Handoff

日期：2026-10-24

## 范围与依据

本 packet 先阅读了项目 `AGENTS.md`、change `proposal.md`、`design.md`、Phase 2 tasks 2.13–2.15、`docs/research/akshare-hk-market-data.md`、现有 `docs/api/hkex-market-data.md`、Docker Compose，以及 Go backend 的 webhook/config/domain contract。未修改 Go 交易逻辑、前端、OpenSpec tasks/state 或跨 Phase 文件。

## 已完成改动

- 新增 `services/akshare/scheduler.py`：
  - 港股代码规范化为五位数字，例如 `700` → `00700`。
  - spot 字段映射为 Go ingestion DTO，固定 `source=akshare`、`data_status=delayed`。
  - history 使用不复权语义，固定 `adjustment=none`、`data_status=unofficial`。
  - payload 固定为 `version=1`、`quotes`、`candles`。
  - HMAC-SHA256 签名格式为 `timestamp.delivery_id.raw_body`，与 Go verifier 对齐。
  - `AKSHARE_POLL_INTERVAL` 默认 900 秒，低于 15 分钟直接拒绝。
  - webhook 最多三次发送，指数退避并加入 jitter；失败不会标记健康。
  - 日志脱敏 secret、token/password 字段和 webhook URL 凭据。
  - 支持配置 spot、标的和 history 回补日期；无配置 history 时不发起回补。
- 新增 `services/akshare/Dockerfile` 与 `services/akshare/requirements.lock`，固定 `akshare==1.19.3`。
- 更新 `deploy/docker-compose.yml`：新增可选 `akshare` profile；服务无 `ports` 映射，仅依赖健康的 backend，并配置容器健康检查。
- 更新 `docs/api/hkex-market-data.md`：记录部署边界、字段、状态、HMAC headers、15 分钟门槛、退避、脱敏、健康检查及 Go 侧事务/广播责任。
- 新增 Python 测试与证据：`services/akshare/tests/test_scheduler.py`、`services/akshare/tests/red-output.txt`、`services/akshare/tests/green-output.txt`。

## TDD 证据

- RED：执行 `python3 -m unittest discover -s services/akshare/tests -v`，在实现前因 `ModuleNotFoundError: No module named 'services.akshare.scheduler'` 失败；完整输出保存在 `services/akshare/tests/red-output.txt`。
- GREEN：执行 `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s services/akshare/tests -v`，8 tests 全部通过；完整输出保存在 `services/akshare/tests/green-output.txt`。
- 覆盖：字段规范化、history 不复权、payload、最短 15 分钟、HMAC、退避计算、日志脱敏、健康状态。

## 命令结果

- `python3 -m unittest discover -s services/akshare/tests -v`：通过，8/8。
- `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s services/akshare/tests -v`：通过，8/8。
- `docker compose -f deploy/docker-compose.yml config`：通过，默认 profile 配置可展开。
- `docker compose --profile akshare -f deploy/docker-compose.yml config`：通过；确认 `akshare` 没有宿主端口映射，使用 backend 内部 webhook。
- `python3 -m py_compile services/akshare/scheduler.py`：未取得通过证据；macOS Python 尝试写入受限的 `~/Library/Caches/com.apple.python/...` 时返回 `PermissionError`。随后以 `PYTHONDONTWRITEBYTECODE=1` 运行完整测试通过。

## 阻塞与准确边界

- Docker daemon 当前不可用：`docker version --format '{{.Server.Version}}'` 返回 `permission denied while trying to connect to the Docker API at unix:///Users/hongxincong/.docker/run/docker.sock`。因此没有伪造 Docker build、启动、真实 AkShare 拉取、webhook→MySQL→REST→认证 WebSocket 链路结果。
- 未等待网络，也未声称真实 AkShare 上游可达；真实依赖安装和真实样本验证需要 Docker daemon/网络恢复后执行。

## 未做项

- 未修改共享 `docs/api/openapi.yaml`：用户限定修改范围只允许本 API 文档，避免跨范围改动；2.13 的共享 OpenAPI 更新仍待后续授权 agent/phase 处理。
- 未修改 OpenSpec tasks/state、Go 交易逻辑、前端或数据库 migration。
- 未完成 2.15 的真实端到端证明、重放/事务回滚/调度停止 stale/重启幂等的 Docker 集成验证；Go 侧已有 webhook contract，但本 packet 未改动其实现。
- 未更新任何任务勾选或 orchestration 状态。

## 停止条件

本 Phase 2 AkShare/deployment packet 已按当前环境完成可验证部分并记录 Docker blocker；不等待外部环境，不继续修改其他 Phase。

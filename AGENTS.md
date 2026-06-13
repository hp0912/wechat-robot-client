# AGENTS.md

## Project Overview

这是微信机器人客户端服务，使用 Go、Gin、GORM、Redis 和 Qdrant。程序入口是 `main.go`，启动流程大致为：

1. `startup.LoadConfig()` 从环境变量或开发模式 `.env` 加载配置。
2. `startup.SetupVars()` 初始化 MySQL 和 Redis。
3. `startup.AutoMigrate()`、`startup.SeedData()` 初始化数据。
4. 注册消息插件、微信机器人、RAG/记忆服务、定时任务和 Agent。
5. `router.RegisterRouter()` 注册 `/api/v1` 下的 HTTP API。

本项目会连接微信协议服务、MySQL、Redis、Qdrant、对象存储、AI 服务、词云服务等外部组件。不要假设本地机器已经具备完整运行环境，除非明确指出，否则不用每次改动都执行单元测试。

## Related Repositories

本仓库只是整套微信机器人系统的一部分。涉及跨仓库任务时，先确认相关仓库的职责和本地路径，不要猜测绝对路径。

- `wechat-robot-client`: 当前仓库，机器人客户端服务。
- `wechat-robot-admin-frontend`: 机器人管理后台前端。
- `wechat-robot-admin-backend`: 机器人管理后台后端。

如果任务需要修改其它仓库：

1. 先检查仓库根目录是否存在 `AGENTS.local.md`。
2. 如果存在，读取其中的 `Local Repository Paths`，并确认目标路径在当前会话的可读/可写范围内。
3. 如果不存在，或者目标路径不在当前会话的工作区权限内，先向用户确认路径或要求从包含相关仓库的工作区启动。

## Repository Layout

- `main.go`: 服务入口和启动编排。
- `router/`: Gin 路由注册。
- `controller/`: HTTP 请求解析、响应返回和控制器层逻辑。
- `service/`: 业务逻辑。新增业务能力时优先放在这里。
- `repository/`: 数据访问封装，避免在控制器中直接操作数据库。
- `model/`: GORM 模型、业务枚举和持久化结构。
- `startup/`: 配置加载、依赖初始化、迁移、种子数据、插件和 Agent 启动。
- `vars/`: 进程级共享依赖和配置。使用时要注意测试隔离。
- `pkg/robot/`: 微信机器人协议客户端和消息结构。
- `pkg/mcp/`: MCP 客户端相关实现。
- `pkg/skills/`: Skills 发现、安装、加载和执行逻辑。
- `pkg/templates/`: 嵌入式模板和静态资源。
- `common_cron/`: 定时任务。
- `plugin/`: 消息插件和扩展能力。
- `.deploy/`: 本地和服务器 Docker Compose 部署配置。

## Common Commands

格式化代码：

```bash
gofmt -w <changed-go-files>
```

运行较快、相对独立的测试：

```bash
go test ./pkg/skills ./pkg/robot ./utils
```

运行指定包测试：

```bash
go test ./path/to/package
```

构建二进制：

```bash
go build ./...
```

本地开发运行需要先准备 `.env` 和外部服务：

```bash
GO_ENV=dev go run .
```

## Configuration

- `GO_ENV=dev` 时会加载根目录 `.env`。
- `.env.example` 是本地开发配置参考。
- 不要提交真实 `.env`、token、API key、数据库密码、证书或本地数据目录。

## Code Style

- 遵循标准 Go 风格，提交前对修改过的 Go 文件运行 `gofmt`。
- 保持现有分层：控制器处理 HTTP，服务层处理业务，仓储层处理数据访问。
- 新接口通常需要同时关注 `controller/`、`service/`、`repository/`、`model/` 和 `router/router.go`。
- 错误要带上下文，避免吞掉外部服务、数据库和文件操作错误。
- 不要在控制器里堆复杂业务逻辑；复杂流程放到 service 或 `pkg/` 的专用组件。
- 避免引入全局状态，确实需要时优先沿用 `vars/` 的现有模式，并注意测试污染。
- 不要随意改动微信协议字段、XML 模板、消息类型枚举和回调结构；这些通常与外部服务协议耦合。

## Database and Migrations

- 模型定义在 `model/`，自动迁移入口在 `startup/migrate.go`。
- 改字段时检查 GORM tag、JSON tag、索引、默认值和已有数据兼容性。
- 机器人实例库和管理后台库是两个连接：`vars.DB` 与 `vars.AdminDB`。确认数据应该写入哪个库。

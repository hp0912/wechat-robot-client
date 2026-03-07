**远程调试**

- 你把 Go 应用打包进 Docker 后，仍然可以在宿主机发起调试。
- 本质是：容器里运行 `dlv --headless`，宿主机 IDE 通过端口连过去。

**最常见做法（推荐）**

- 编译调试版二进制（关闭优化）：
  - `go build -gcflags="all=-N -l" -o app ./cmd/xxx`
- 容器里启动 Delve：
  - `dlv exec ./app --headless --listen=:40000 --api-version=2 --accept-multiclient`
- 映射端口到宿主机：
  - `-p 40000:40000`
- 宿主机 IDE 用“Remote”模式连接 `127.0.0.1:40000`。

**容器权限坑（最容易卡住）**

- 需要允许 `ptrace`，否则经常报 `operation not permitted`。
- 运行容器时加：
  - `--cap-add=SYS_PTRACE`
  - 常见还要 `--security-opt seccomp=unconfined`
- 在某些环境（K8s、受限平台）会被策略拦截，需要平台侧放行。

**IDE 配置要点**

- VS Code（Go 插件）`launch.json` 里用 `request: "attach"` + `mode: "remote"` + `host/port`。
- GoLand 选 `Go Remote`，填主机和端口即可。
- 断点能否命中强依赖“源码路径映射”和“编译参数（-N -l）”。

**生产环境建议**

- 不建议长期暴露调试端口（有安全风险）。
- 远程调试会影响性能，适合临时排障。
- 线上优先考虑 `pprof`、结构化日志、trace；远程 debugger 用于疑难问题短时介入。

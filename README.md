# 🛡️ Kube-Sentinel

**Kubernetes 集群自愈监控系统** — 基于 Informer（List-Watch）机制的实时事件监控与告警

## 核心特性

- **Informer 事件驱动监控**：使用 SharedInformerFactory 监听 Pod 生命周期事件，替代低效的轮询方案
- **智能重启检测**：自动检测容器重启次数超过阈值的异常 Pod，支持 CrashLoopBackOff、ImagePullBackOff 等状态识别
- **Prometheus 集成**：实时查询 Pod CPU/内存使用率，发现异常 Pod
- **结构化日志**：基于 zap 的高性能日志，支持 Debug/Info/Warn/Error 级别
- **优雅关闭**：监听 OS 信号（SIGINT/SIGTERM），通知所有 goroutine 安全退出
- **生产级配置**：通过环境变量或配置文件管理所有参数

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Kube-Sentinel                          │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  Pod Watcher  │    │  Prometheus  │    │   Alerter    │  │
│  │  (Informer)   │    │   Client     │    │  (微信/钉钉) │  │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘  │
│         │                   │                    │          │
│  ┌──────▼───────────────────▼────────────────────▼───────┐  │
│  │                    Event Bus                           │  │
│  └───────────────────────┬───────────────────────────────┘  │
│                          │                                  │
│  ┌───────────────────────▼───────────────────────────────┐  │
│  │              Structured Logger (zap)                   │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                          │
         ▼                          ▼
  K8s API Server              Prometheus
  (List-Watch)                (PromQL)
```

## 快速开始

### 前置条件

- Go 1.20+
- Minikube 或 K8s 集群
- （可选）Prometheus

### 本地运行

```bash
# 1. 编译
go build -o bin/sentinel.exe ./cmd/sentinel/

# 2. 运行（自动使用 ~/.kube/config）
./bin/sentinel.exe

# 3. 启用 Prometheus 查询
PROMETHEUS_ENABLED=true PROMETHEUS_URL=http://localhost:9090 ./bin/sentinel.exe
```

### 测试场景

```bash
# 制造 ImagePullBackOff 事件
kubectl run broken --image=nginx:nonexistent

# 制造容器重启
kubectl run crasher --image=busybox -- sh -c "exit 1"

# 正常创建和删除
kubectl run nginx --image=nginx
kubectl delete pod nginx
```

## 配置参数

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `KUBECONFIG` | `~/.kube/config` | kubeconfig 路径 |
| `WATCH_NAMESPACE` | `""` (所有) | 监听的命名空间 |
| `RESTART_THRESHOLD` | `3` | 重启告警阈值 |
| `RESYNC_INTERVAL` | `30s` | Informer 全量同步间隔 |
| `PROMETHEUS_ENABLED` | `false` | 是否启用 Prometheus |
| `PROMETHEUS_URL` | `http://localhost:9090` | Prometheus 地址 |
| `QUERY_INTERVAL` | `60s` | Prometheus 查询间隔 |
| `LOG_LEVEL` | `info` | 日志级别 |

## 项目结构

```
kube-sentinel/
├── cmd/sentinel/main.go           # 应用入口
├── internal/
│   ├── config/app.go              # 配置管理
│   ├── k8s/
│   │   ├── client.go              # K8s Clientset 初始化
│   │   └── config.go              # kubeconfig 加载
│   ├── monitor/
│   │   ├── pod_watcher.go         # Informer Pod 监控（核心）
│   │   └── types.go               # 数据类型定义
│   └── prometheus/
│       ├── client.go              # Prometheus 客户端
│       └── query.go               # PromQL 查询
├── pkg/logger/logger.go           # 结构化日志
├── deployments/
│   ├── Dockerfile                 # 多阶段构建
│   ├── deployment.yaml            # K8s 部署
│   └── rbac.yaml                  # RBAC 权限
├── Makefile
├── go.mod
└── go.sum
```

## 面试要点

### 为什么用 Informer 而不是轮询？

Informer 建立 List-Watch 通道：首次从 API Server List 全量对象写入本地缓存，后续 Watch 会推送增量事件。所有事件在内存中聚合，读写不再频繁访问 API Server，从而大幅降低延迟和资源消耗。

### WaitForCacheSync 的作用？

在处理事件前调用 `WaitForCacheSync` 确保本地 SharedInformer 缓存已经与 API Server 同步完成，否则可能因为缓存未准备好而漏掉 Pod 或误判 Pod 状态。

### 如何处理 DeletedFinalStateUnknown？

当 Informer 因网络问题丢失删除事件时，会收到一个 `DeletedFinalStateUnknown` Tombstone。正确做法是从 Tombstone 中提取原始对象，而不是直接断言类型为 `*v1.Pod`，否则会触发 panic。

## License

MIT


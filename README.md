# Webook

以「内容社区」为业务载体的 Go 微服务实战项目。目标不是堆业务 CRUD，而是把**分层设计、依赖注入、缓存一致性、事件驱动、可靠性、数据迁移、服务治理与可观测性**串成一条可复现的学习路径。

> 适合作为面试叙事：单体 SRD → 中间件与缓存 → 拆微服务 → 事件与可靠性 → 双写迁移 → 服务发现与治理 → 可观测与部署。

---

## 项目定位

| 维度 | 说明 |
|------|------|
| 业务形态 | 用户 / 文章 / 互动（读赞藏）/ 关注 / Feed / 评论 / 标签 / 搜索 / 支付打赏 |
| 工程形态 | 主站 BFF（Gin HTTP）+ 多 gRPC 微服务 + 共享 `pkg` |
| 学习目标 | 理解「为什么这样拆、这样写、这样兜底」，而不只是「能跑」 |

前端：`webook-fe`（Next.js）。本地依赖见 `webook/docker-compose.yaml`。

---

## 总体架构

```text
                    ┌─────────────────────────────────────┐
                    │  Client / webook-fe                 │
                    └─────────────────┬───────────────────┘
                                      │ HTTP
                    ┌─────────────────▼───────────────────┐
                    │  BFF  webook/main.go                │
                    │  Gin · JWT · 限流 · Metric · Trace  │
                    │  + Kafka Consumer · Cron · Job      │
                    └─┬───────┬───────┬───────┬───────────┘
                      │ gRPC  │       │       │
         ┌────────────┼───────┼───────┼───────┼────────────┐
         ▼            ▼       ▼       ▼       ▼            ▼
   interactive    follow    feed   payment  reward     search / tag
   comment        account   ...    (本地消息)            comment ...
         │            │       │       │                    │
         └────────────┴───────┴───────┴────────────────────┘
                              │
              MySQL · Redis · Kafka · Etcd · ES · Mongo
              Canal(binlog) · Prometheus · Zipkin · ELK
```

**设计取舍：**

- **BFF 聚合读路径**：HTTP 面向前端；重逻辑与独立演进的能力下沉为 gRPC 服务。
- **Consumer / Cron / HTTP / gRPC 同级**：都是 App 级入口，生命周期统一管理（见 `startup/App`、`pkg/wego.App`）。
- **横切能力进中间件与 `pkg`**：限流、熔断、双写、Canal、注册发现不散落在业务 Handler 里。

---

## 目录结构

```text
webook/                          # monorepo 根（go.mod）
├── webook-fe/                   # 前端
└── webook/                      # 后端
    ├── main.go                  # BFF：Web + Consumers + Cron + Scheduler
    ├── startup/                 # Wire 装配、App 聚合
    ├── ioc/                     # DB / Redis / Kafka / Etcd / OTEL / 中间件
    ├── internal/                # 主站领域（用户、文章、排行榜等）
    │   ├── domain/
    │   ├── service/
    │   ├── repository/{dao,cache}/
    │   ├── web/{middleware,jwt}/
    │   ├── events/
    │   └── job/
    ├── account/ comment/ follow/ feed/
    ├── interactive/ payment/ reward/ search/ tag/
    │   └── domain · service · repository · grpc · ioc · events · wire · main
    ├── api/proto/               # 跨服务契约
    ├── pkg/                     # grpcx · migrator · gormx · canalx · wego
    ├── notes/                   # 学习笔记（设计动机与面试点）
    ├── docker-compose.yaml
    └── k8s-*.yaml
```

---

## 分层：Service · Repository · DAO

参考 DDD 思路的 SRD 分层（笔记见 `webook/notes/remark.md`）：

| 层 | 职责 | 学习点 |
|----|------|--------|
| **domain** | 领域对象 | 业务语义与存储模型解耦 |
| **service** | 领域服务 | 用例编排、事务边界、跨仓储逻辑 |
| **repository** | 仓储 | 对上屏蔽「缓存 + DB + 外部依赖」 |
| **dao / cache** | 持久化与缓存 | SQL / Redis / ES 细节下沉 |
| **web / grpc** | 接入层 | DTO、鉴权、错误码，尽量薄 |
| **ioc / wire** | 组装 | 依赖注入，避免包级全局变量 |
| **events / job** | 异步入口 | 与 Web/gRPC 同级，便于扩展 |

**依赖三原则（工程习惯）：**

1. A 依赖 B → B 是**接口**（面向抽象）
2. A 依赖 B → B 是 A 的**字段**（规避包变量 / 包方法）
3. A 不创建 B → **外部注入**（依赖反转，Wire 落地）

微服务目录基本复用同一模板，便于从单体「抄一层」拆出去（`interactive` 仍复用部分 `internal` 实现，可看作拆分过渡态）。

---

## 微服务职责

| 模块 | 职责 |
|------|------|
| **主站 BFF** | 用户 / 文章 / OAuth / 排行榜 / 打赏代理；挂载 Consumer、Cron、MySQL 抢占调度 |
| **interactive** | 阅读 / 点赞 / 收藏计数；Kafka 消费阅读事件；双写迁移 Admin |
| **follow** | 关注关系、粉丝 / 关注列表与统计 |
| **feed** | Feed 推 / 拉事件；按事件类型 Handler；大 V 阈值切换 |
| **comment** | 评论；含限流与异步创建变体 |
| **tag** | 用户标签 |
| **search** | ES 同步与搜索；消费文章 / 用户同步事件 |
| **payment** | 微信 Native 支付；回调 Web；本地消息表补偿 |
| **reward** | 打赏订单，对接支付与账户 |
| **account** | 账户入账 / 积分 |

契约：`webook/api/proto/`，生成命令见根目录 `Makefile` 的 `grpc` target。

---

## 架构亮点（学习主线）

### 1. Wire 依赖注入

`startup/wire.go` 与各服务 `wire.go` 用 Provider Set + `wire.Bind` 把接口实现绑死在编译期。

**学到什么：** 可测、可替换不是口号——Mock 测 Service 时，Repository 是接口；换缓存实现时，Service 不用改。

### 2. Middleware = AOP

HTTP 链路上集中处理：AccessLog、Prometheus、OTEL（`otelgin`）、CORS、JWT 登录、滑动窗口限流（`ioc/web.go`、`internal/web/middleware`）。

gRPC 侧对称：限流拦截器、熔断（`pkg/grpcx/interceptors`）、自定义 WRR 负载均衡（`pkg/grpcx/balancer/wrr`）。

**学到什么：** 横切关注点与业务 Handler 分离；治理能力应下沉到框架包，而不是每个接口手写一遍。

### 3. 互动计数与缓存

- DB：`Cnt = Cnt + 1` / upsert，避免「先读后写」并发丢更新（`notes/interactive.md`）
- Redis：Hash 多字段 + Lua；`Incr*IfPresent`（仅缓存命中才加）
- 异步：文章阅读发 Kafka → Interactive 消费加阅读数

**学到什么：** 缓存策略要按读写比与一致性要求选型；「一次 HGETALL vs 三次 GET」是网络往返与原子性的权衡，不是单纯「哪个更省 key」。

### 4. Feed 推拉结合

- **Pull**：作者发件箱（适合大 V，避免写扩散）
- **Push**：粉丝收件箱（适合普通用户，读路径简单）
- 文章事件按粉丝数阈值切换（`feed/service/article_event.go`）；读时合并 Push + Pull 再排序

**学到什么：** 经典热点用户问题——写扩散 vs 读聚合，用阈值做工程折中，而不是二选一教条。

### 5. 支付可靠性：本地消息表

`payment`：业务与发消息同事务写本地消息 → Job 扫描 Init 重发 → 超时 MarkFailed；消费侧要求幂等。微信回调 + 主动 `SyncWechatInfo` 双通道对账。

**学到什么：** 「发 MQ 一定成功」不可信；本地消息 / Outbox 是跨系统最终一致性的标准套路之一。

### 6. 无停机数据迁移

- 双写连接池：`SrcOnly → SrcFirst → DstFirst → DstOnly`（`pkg/gormx/connpool`）
- 全量校验 / 修复：`pkg/migrator/{validator,fixer,scheduler}`
- 增量：Canal 订阅 binlog → Kafka → 消费修复

**学到什么：** 改存储 / 拆库不是「停机 dump」，而是**模式切换 + 校验 + 增量兜底**的完整故事，面试可完整展开。

### 7. 任务调度与分布式锁

- Cron 排行榜任务 + Redis 分布式锁 AutoRefresh（`internal/job/ranking_job.go`）
- 基于 MySQL 的抢占式 Scheduler（`main.go` 中 `app.Scheduler()`）

**学到什么：** 多实例下「只跑一份」要靠锁或抢占；长任务要考虑锁续租与优雅退出超时。

### 8. 可观测性三件套

| 信号 | 实现 |
|------|------|
| Metrics | Prometheus `:8081/metrics`，Gin 业务码 / 延迟直方图 |
| Tracing | OpenTelemetry → Zipkin |
| Logging | Filebeat → Logstash → ES → Kibana |

本地全家桶：`webook/docker-compose.yaml`（MySQL binlog、Redis、Etcd、Mongo、Kafka、ES、Prometheus、Grafana、Zipkin、Canal 等）。

### 9. 部署练习

Docker 镜像构建 + `k8s-*.yaml`（Deployment / Service / Ingress / MySQL / Redis），把「本地能跑」推到「集群里能挂」。

---

## 技术栈

| 类别 | 选型 |
|------|------|
| 语言 / 模块 | Go，Google Wire |
| HTTP / RPC | Gin，gRPC + Protobuf |
| 数据 | MySQL（GORM）、Redis、MongoDB、Elasticsearch |
| 消息 / 同步 | Kafka（Sarama）、Canal |
| 注册发现 | Etcd |
| 可观测 | Prometheus、OTEL/Zipkin、ELK |
| 支付 | 微信支付 API v3 |
| 配置 | Viper（本地 / 远程） |

---

## 快速开始

### 基础设施

```bash
cd webook
docker compose up -d
```

### 主站 BFF

```bash
# 在仓库根目录
go run ./webook
# HTTP :8080  ·  Metrics :8081
```

### 前端

```bash
cd webook-fe
npm run dev
# http://localhost:3000
```

### 生成 Mock / Proto

```bash
make mock
make grpc
```

各微服务入口：`webook/{interactive,payment,feed,...}/main.go`，按需单独启动。

---

## 建议学习顺序

1. **分层与 DI**：`internal/` + `startup/wire.go` + `notes/remark.md`
2. **登录与中间件**：JWT Middleware、限流、Prometheus / OTEL
3. **互动与缓存**：`notes/interactive.md` + repository/cache + Kafka 阅读事件
4. **拆服务**：`interactive` / `follow` gRPC + Etcd 客户端（`ioc/intr.go`）
5. **Feed 推拉**：`feed/service/article_event.go`
6. **支付与本地消息**：`payment/job/local_msg_job.go`
7. **双写与 Canal**：`pkg/gormx/connpool`、`pkg/migrator`、`pkg/canalx`
8. **治理与可观测**：`pkg/grpcx`、compose 中 Prometheus / Zipkin / ELK
9. **K8s**：`k8s-webook-*.yaml`

配合 `webook/notes/` 里的面试题与设计笔记，把「代码路径」翻译成「能讲清楚的决策」。

---

## 设计原则小结

```text
业务逻辑在 Service，存储细节在 DAO/Cache
接入层尽量薄；异步入口与 HTTP/gRPC 同级
横切用 AOP；依赖用接口 + 注入
一致性优先想清楚：强一致 / 最终一致 / 补偿 / 幂等
热点与迁移用工程折中，而不是银弹组件
可观测与部署是架构的一部分，不是事后补丁
```

---

## License

个人学习 / 实战练习项目。

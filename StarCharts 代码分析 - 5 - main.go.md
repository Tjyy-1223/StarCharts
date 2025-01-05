## main.go 代码分析

```go
package main

import (
	"B1-StarCharts/config"
	"B1-StarCharts/controller"
	"B1-StarCharts/internal/cache"
	"B1-StarCharts/internal/github"
	"embed"
	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)

//go:embed static/*
var static embed.FS

var version = "devel"

func main() {
	log.SetHandler(text.New(os.Stderr))
	// log.SetLevel(log.DebugLevel)
	config := config.Get()
	ctx := log.WithField("listen", config.Listen)
	options, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		log.WithError(err).Fatal("invalid redis_url")
	}

	redis := redis.NewClient(options)
	cache := cache.New(redis)
	defer cache.Close()
	github := github.New(config, cache)

	r := mux.NewRouter()
	r.Path("/").Methods(http.MethodGet).Handler(controller.Index(static, version))
	r.Path("/").Methods(http.MethodPost).Handler(controller.HandleForm())
	r.PathPrefix("/static/").Methods(http.MethodGet).Handler(http.FileServer(http.FS(static)))
	r.Path("/{owner}/{repo}.svg").
		Methods(http.MethodGet).
		Handler(controller.GetRepoChart(github, cache))
	r.Path("/{owner}/{repo}").
		Methods(http.MethodGet).
		Handler(controller.GetRepo(static, github, cache, version))

	// generic metrics
	requestCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "total requests",
	}, []string{"code", "method"})
	responseObserver := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "starcharts",
		Subsystem: "http",
		Name:      "responses",
		Help:      "response times and counts",
	}, []string{"code", "method"})

	r.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())
	srv := &http.Server{
		Handler: httplog.New(
			promhttp.InstrumentHandlerDuration(
				responseObserver,
				promhttp.InstrumentHandlerCounter(
					requestCounter,
					r,
				),
			),
		),
		Addr:         config.Listen,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	ctx.Info("starting up")
	ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
}
```

这段代码是一个使用 Go 编写的 HTTP 服务器程序，功能上实现了以下几个主要任务：

1. **读取配置**：从配置文件或环境变量中获取相关的配置信息。
2. **Redis 缓存**：初始化一个 Redis 客户端用于缓存操作。
3. **GitHub API 集成**：与 GitHub 交互，获取仓库信息。
4. **嵌入静态资源**：将静态文件嵌入到程序中，使其作为嵌入文件在程序中使用。
5. **路由设置**：配置了不同的 HTTP 路由与对应的处理函数。
6. **Prometheus 指标**：集成了 Prometheus，用于监控 HTTP 请求与响应。
7. **日志处理**：使用 `apex/log` 进行日志记录。
8. **启动 HTTP 服务**：最终启动了一个 HTTP 服务来监听请求并处理。

下面是详细解释：

### 1. **导入的包**

```go
import (
	"B1-StarCharts/config"
	"B1-StarCharts/controller"
	"B1-StarCharts/internal/cache"
	"B1-StarCharts/internal/github"
	"embed"
	"github.com/apex/httplog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)
```

这些是该程序所依赖的第三方库和自定义包：

- `config`, `controller`, `cache`, `github` 是自定义的包，负责配置、控制器逻辑、缓存和 GitHub API 的集成。
- `embed`：Go 1.16 引入的包，用于将文件嵌入到程序中。
- `httplog` 和 `log`：用于 HTTP 请求日志的记录。
- `redis`：Go 的 Redis 客户端。
- `mux`：一个流行的 HTTP 路由库。
- `prometheus` 和 `promhttp`：用于集成 Prometheus 进行监控。

### 2. **嵌入静态资源**

```go
//go:embed static/*
var static embed.FS
```

- 使用 `embed` 包将 `static/` 目录下的所有文件嵌入到程序中。这样可以直接从内存中提供静态文件，而不需要将其单独存储在磁盘上。

### 3. **读取配置和初始化 Redis 客户端**

```go
config := config.Get()
ctx := log.WithField("listen", config.Listen)
options, err := redis.ParseURL(config.RedisURL)
if err != nil {
	log.WithError(err).Fatal("invalid redis_url")
}

redis := redis.NewClient(options)
cache := cache.New(redis)
defer cache.Close()
```

- 从配置中获取服务监听地址和 Redis URL。
- 初始化 Redis 客户端，连接到 Redis 服务器并创建缓存实例。
- 使用 `defer` 保证在程序结束时关闭 Redis 连接。

### 4. **创建 GitHub 客户端**

```go
github := github.New(config, cache)
```

- 创建一个 GitHub 客户端实例，传入配置和缓存实例，用于与 GitHub 进行交互，可能是获取仓库信息、处理相关请求等。

### 5. **设置路由**

```go
r := mux.NewRouter()
r.Path("/").Methods(http.MethodGet).Handler(controller.Index(static, version))
r.Path("/").Methods(http.MethodPost).Handler(controller.HandleForm())
r.PathPrefix("/static/").Methods(http.MethodGet).Handler(http.FileServer(http.FS(static)))
r.Path("/{owner}/{repo}.svg").
	Methods(http.MethodGet).
	Handler(controller.GetRepoChart(github, cache))
r.Path("/{owner}/{repo}").
	Methods(http.MethodGet).
	Handler(controller.GetRepo(static, github, cache, version))
```

- 使用 `mux.NewRouter()` 创建了一个路由器，并设置了多个路由处理函数。
- `/` 路径的 GET 和 POST 请求分别由 `controller.Index` 和 `controller.HandleForm` 处理。
- `/static/` 路径用于提供静态文件，使用嵌入的静态文件系统。
- `/{owner}/{repo}.svg` 和 `/{owner}/{repo}` 是 GitHub 仓库相关的路径，分别由 `controller.GetRepoChart` 和 `controller.GetRepo` 处理，前者返回一个 SVG 格式的仓库图表，后者返回仓库的详细信息。

### 6. **Prometheus 监控设置**

```go
// generic metrics
requestCounter := promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "http",
	Name:      "requests_total",
	Help:      "total requests",
}, []string{"code", "method"})
responseObserver := promauto.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "starcharts",
	Subsystem: "http",
	Name:      "responses",
	Help:      "response times and counts",
}, []string{"code", "method"})
```

- 定义了两个 Prometheus 指标：
  - `requestCounter`：计数器，用于统计不同 HTTP 状态码和方法（GET, POST 等）的请求总数。
  - `responseObserver`：一个摘要指标，用于跟踪响应时间。

```go
r.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler())
```

- 设置了 `/metrics` 路径，用于暴露 Prometheus 格式的监控指标。

### 7. **启动 HTTP 服务器**

```go
srv := &http.Server{
	Handler: httplog.New(
		promhttp.InstrumentHandlerDuration(
			responseObserver,
			promhttp.InstrumentHandlerCounter(
				requestCounter,
				r,
			),
		),
	),
	Addr:         config.Listen,
	WriteTimeout: 60 * time.Second,
	ReadTimeout:  60 * time.Second,
}
ctx.Info("starting up")
ctx.WithError(srv.ListenAndServe()).Error("failed to start up server")
```

- 创建并配置 HTTP 服务器，设置请求处理器为 `httplog.New`，它结合了 Prometheus 监控和 HTTP 请求日志。
- 配置了请求的超时时间（60 秒）和监听地址。
- 调用 `srv.ListenAndServe()` 启动 HTTP 服务并开始监听请求。
- 在启动时记录日志，若启动失败则记录错误。

### 总结

这段代码实现了一个基础的 Go Web 服务，它集成了 Prometheus 用于监控，Redis 用于缓存，静态文件嵌入功能，GitHub 仓库相关的处理功能，并且具有 HTTP 请求日志和 Prometheus 指标导出功能。这个服务的主要功能是展示 GitHub 仓库的图表、信息，并提供监控接口。


# StarCharts 代码分析 - 1 - config pkg

### 1.1 代码分析

config 用来读取环境配置，其代码如下：

```go
package config

import (
	"github.com/apex/log"
	"github.com/caarlos0/env/v6"
)

// Config configuration.
type Config struct {
	RedisURL              string   `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	GitHubTokens          []string `env:"GITHUB_TOKENS"`
	GitHubPageSize        int      `env:"GITHUB_PAGE_SIZE" envDefault:"100"`
	GitHubMaxRateUsagePct int      `env:"GITHUB_MAX_RATE_LIMIT_USAGE" envDefault:"80"`
	Listen                string   `env:"LISTEN" envDefault:"127.0.0.1:3000"`
}

// Get the current Config.
func Get() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	return
}
```

`Config` 是一个结构体，用于存储从环境变量中加载的配置项。每个字段都用结构体标签（`env` 标签）进行注解，告诉 `env` 包从环境变量中获取哪些配置，并在没有找到该环境变量时使用默认值（通过 `envDefault` 标签指定）。下面是每个字段的说明：

- `RedisURL`：Redis 服务器的 URL，默认值为 `redis://localhost:6379`。
- `GitHubTokens`：GitHub 的令牌，类型为 `[]string`，即一个字符串切片，用于存储多个令牌。没有默认值，必须通过环境变量提供。
- `GitHubPageSize`：每次查询 GitHub API 时，返回的最大结果数。默认值为 `100`。
- `GitHubMaxRateUsagePct`：GitHub API 请求的最大速率限制使用百分比，默认值为 `80`，表示最大使用 80% 的速率限制。
- `Listen`：服务器监听的地址和端口，默认值为 `127.0.0.1:3000`。

而对于获取配置的函数：

```go
func Get() (cfg Config) {
	if err := env.Parse(&cfg); err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	return
}
```

+ `Get` 函数用于获取并返回一个 `Config` 配置对象。
+ `env.Parse(&cfg)`：这个方法会扫描环境变量并将其解析到 `Config` 结构体的字段中。如果某个字段的环境变量存在，它会把环境变量的值映射到字段上。如果环境变量没有设置并且该字段有默认值，则使用默认值。
+ 如果解析过程中发生错误（例如缺少必需的环境变量），则会触发错误处理：`log.WithError(err).Fatal("failed to load config")`。这行代码会记录一个错误日志，并且通过 `Fatal` 终止程序。
+ 如果没有发生错误，`cfg`（类型为 `Config`）将被返回，包含从环境变量中解析到的配置数据。

这样，你可以通过设置环境变量来动态调整程序的配置，而无需修改代码。



### 1.2 结构体标签的使用

在 Go 语言中，**结构体标签**是用于为结构体的字段附加元数据的机制，这些元数据可以通过反射来访问。结构体标签通常用于指定额外的配置信息或行为，供包或框架使用。

在给出的代码中，结构体标签是 `env` 和 `envDefault`，它们的作用是告诉 `env` 包如何从环境变量中读取数据并进行映射。具体来说，`env` 标签用于指定对应的环境变量名称，而 `envDefault` 标签则用于指定默认值，当环境变量不存在时使用。

**综上，结构体标签作用如下：**

1. **映射环境变量到字段**：指定从哪个环境变量获取配置值。

2. **设置默认值**：如果环境变量没有设置，使用默认值来初始化字段。

`env:"REDIS_URL"` 这样的标签告诉 `env` 包：**从环境变量中获取值，并将它映射到结构体字段**。例如，字段 `RedisURL` 会从环境变量 `REDIS_URL` 中获取值。

#### 1.2.1 `env` 标签

`env:"REDIS_URL"` 这样的标签告诉 `env` 包：**从环境变量中获取值，并将它映射到结构体字段**。例如，字段 `RedisURL` 会从环境变量 `REDIS_URL` 中获取值。

**举例：**

- 假设你设置了环境变量：

  ```
  bash
  
  
  复制代码
  export REDIS_URL="redis://some-other-host:6379"
  ```

- 那么，`RedisURL` 字段的值将被解析为 `redis://some-other-host:6379`。

如果环境变量 `REDIS_URL` 没有设置，则会使用字段的默认值 `redis://localhost:6379`（通过 `envDefault` 标签）。

#### 1.2.2 `envDefault` 标签

`envDefault:"<default_value>"` 标签用于为结构体字段提供默认值。**当对应的环境变量没有被设置时，Go 程序将使用这个默认值**。

**举例：**

- `GitHubPageSize` 字段有标签 `envDefault:"100"`，这表示如果环境变量 `GITHUB_PAGE_SIZE` 没有被设置，`GitHubPageSize` 字段将使用默认值 `100`。
- 如果你没有设置环境变量 `GITHUB_PAGE_SIZE`，那么 `GitHubPageSize` 就会被初始化为 `100`。

#### 1.2.3 完整示例

假设你的环境变量如下所示：

```bash
export REDIS_URL="redis://localhost:6380"
export GITHUB_TOKENS="token1,token2"
```

那么当你调用 `Get()` 函数时，结构体 `Config` 会被填充如下：

```go
Config{
    RedisURL:              "redis://localhost:6380",  // 从环境变量 REDIS_URL 获取
    GitHubTokens:          []string{"token1", "token2"}, // 从环境变量 GITHUB_TOKENS 获取
    GitHubPageSize:        100,  // 使用默认值 100（因为环境变量 GITHUB_PAGE_SIZE 没有设置）
    GitHubMaxRateUsagePct: 80,   // 使用默认值 80（因为环境变量 GITHUB_MAX_RATE_LIMIT_USAGE 没有设置）
    Listen:                "127.0.0.1:3000",  // 使用默认值 127.0.0.1:3000（因为环境变量 LISTEN 没有设置）
}
```

结构体标签的作用主要是通过反射提供额外的信息，以便某些包（如 `github.com/caarlos0/env`）能够解析环境变量，并将其映射到结构体字段中。具体来说：

- `env` 标签用于指定环境变量名称，告诉 `env` 包从哪个环境变量获取配置。
- `envDefault` 标签则指定默认值，当环境变量未设置时使用该值。

这种方式简化了配置的管理，让程序可以通过环境变量来动态调整配置，而无需修改代码或重启程序。
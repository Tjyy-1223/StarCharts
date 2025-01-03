## StarCharts 代码分析 - 2 - controller pkg

### 1 cache/cache.go

该模块中代码如下：

```go
package cache

import (
	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

// nolint: gochecknoglobals
var cacheGets = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache",
		Name:      "gets_total",
		Help:      "Total number of successful cache gets",
	},
)

// nolint: gochecknoglobals
var cachePuts = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache",
		Name:      "puts_total",
		Help:      "Total number of successful cache puts",
	},
)

// nolint: gochecknoglobals
var cacheDeletes = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache",
		Name:      "deletes_total",
		Help:      "Total number of successful cache deletes",
	},
)

// nolint: gochecknoinits
func init() {
	prometheus.MustRegister(cacheGets, cachePuts, cacheDeletes)
}

// Redis cache
type Redis struct {
	redis *redis.Client
	codec *rediscache.Codec
}

// New redis cache.
func New(redis *redis.Client) *Redis {
	codec := &rediscache.Codec{
		Redis: redis,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &Redis{
		redis: redis,
		codec: codec,
	}
}

// Close connections
func (c *Redis) Close() error {
	return c.redis.Close()
}

// Get from cache by key.
func (c *Redis) Get(key string, result interface{}) error {
	if err := c.codec.Get(key, result); err != nil{
		return err
	}
	cacheGets.Inc()
	return nil
}

// Put on cache.
func (c *Redis) Put(key string, obj interface{}) error {
	if err := c.codec.Set(&rediscache.Item{
		Key: key,
		Object: obj,
	}); err != nil{
		return err
	}
	cachePuts.Inc()
	return nil
}

// Delete from cache.
func (c *Redis) Delete(key string) error {
	if err := c.codec.Delete(key); err != nil{
		return err
	}
	cacheDeletes.Inc()
	return nil
}
```

这段代码是一个简单的 Redis 缓存封装，它使用了 `github.com/go-redis/redis` 包来与 Redis 进行交互，并利用 `github.com/go-redis/cache` 包实现了基于 Redis 的缓存机制。代码还结合了 Prometheus 来监控缓存操作的统计数据（例如：缓存的 `get`、`put` 和 `delete` 操作次数）。

#### 1.1  **Prometheus 计数器**

```go
var cacheGets = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "starcharts",
		Subsystem: "cache",
		Name:      "gets_total",
		Help:      "Total number of successful cache gets",
	},
)
```

这里创建了三个 **Prometheus Counter**，用于记录缓存操作的次数：

- **`cacheGets`**：记录成功的缓存 `get` 操作的次数。
- **`cachePuts`**：记录成功的缓存 `put` 操作的次数。
- **`cacheDeletes`**：记录成功的缓存 `delete` 操作的次数。

这三个计数器将用于 Prometheus 数据采集和监控，以便能够在 Prometheus 中查看缓存的使用情况。

#### 1.2. **`init` 函数**

```go
func init() {
	prometheus.MustRegister(cacheGets, cachePuts, cacheDeletes)
}
```

- `init` 函数在程序启动时会自动执行。它将三个 Prometheus 计数器（`cacheGets`, `cachePuts`, `cacheDeletes`）注册到 Prometheus 的默认注册表中。这样，Prometheus 就可以在运行时收集这三个计数器的数据

#### 1.3. **Redis 缓存结构体**

```go
type Redis struct {
	redis *redis.Client
	codec *rediscache.Codec
}
```

`Redis` 结构体封装了一个 `*redis.Client` 和 `*rediscache.Codec`。

- `redis`: 这个字段是 `github.com/go-redis/redis` 提供的 Redis 客户端，负责与 Redis 服务器进行连接和执行基本的 Redis 操作。
- `codec`: 这个字段是 `github.com/go-redis/cache` 提供的缓存编码器，封装了数据的序列化和反序列化操作，支持缓存数据的存取。

#### 1.4. **`New` 函数**

```fo
func New(redis *redis.Client) *Redis {
	codec := &rediscache.Codec{
		Redis: redis,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	return &Redis{
		redis: redis,
		codec: codec,
	}
}
```

- `New` 函数创建并返回一个 `*Redis` 实例，封装了 Redis 客户端和缓存编码器（`codec`）。
- `codec` 的 `Marshal` 和 `Unmarshal` 方法使用了 `msgpack` 进行数据的序列化和反序列化。`msgpack` 是一种二进制的序列化格式，通常比 JSON 更高效。
- 通过 `&Redis{}`，`New` 返回了一个包含 `redis` 客户端和 `codec` 编码器的缓存对象。

#### 1.5. **`Close` 函数**

```go
func (c *Redis) Close() error {
	return c.redis.Close()
}
```

- `Close` 方法用于关闭 Redis 客户端连接。它直接调用了 `*redis.Client` 的 `Close()` 方法来关闭连接。

#### 1.6. **`Get` 函数**

```go
func (c *Redis) Get(key string, result interface{}) error {
	if err := c.codec.Get(key, result); err != nil {
		return err
	}
	cacheGets.Inc()
	return nil
}
```

- `Get` 方法从 Redis 缓存中获取一个值。它调用了 `codec.Get` 方法，该方法会尝试根据指定的 `key` 从 Redis 中检索数据，并将数据反序列化到 `result` 变量中。
- 如果 `Get` 操作成功，会调用 `cacheGets.Inc()` 增加 Prometheus 中的 `gets_total` 计数器，以统计成功的缓存读取操作。

与此类似：

+ `Put` 方法将一个对象存入 Redis 缓存。它通过 `codec.Set` 方法将数据序列化并保存到 Redis 中。
+ 如果 `Put` 操作成功，会调用 `cachePuts.Inc()` 增加 Prometheus 中的 `puts_total` 计数器，以统计成功的缓存写入操作。

+ `Delete` 方法从 Redis 缓存中删除一个键值对。它通过 `codec.Delete` 方法执行 Redis 删除操作。
+ 如果 `Delete` 操作成功，会调用 `cacheDeletes.Inc()` 增加 Prometheus 中的 `deletes_total` 计数器，以统计成功的缓存删除操作。

#### 1.7 总结

1. **缓存实现**：该代码实现了一个简单的 **Redis 缓存** 系统，封装了缓存的读取、写入和删除操作。它依赖于 `github.com/go-redis/redis` 和 `github.com/go-redis/cache` 库来进行 Redis 的基本操作和缓存管理。

2. **序列化**：使用 `msgpack` 作为序列化和反序列化工具。相比 JSON，`msgpack` 是一种更高效的二进制格式，适合用于 Redis 的缓存数据。

3. **Prometheus 监控**：使用 Prometheus 监控缓存操作，包括缓存的获取 (`get`)、存储 (`put`) 和删除 (`delete`) 次数。每当这些操作成功时，相关的计数器会自增。

4. **封装**：`Redis` 类型将 Redis 客户端和缓存操作封装在一个结构体中，提供了简单的 `Get`、`Put` 和 `Delete` 方法接口，让使用者更容易管理缓存操作。

5. **线程安全**：由于 Redis 客户端和 `rediscache.Codec` 本身是线程安全的，所以该实现是可以在并发环境下使用的。

6. **适用场景**：该实现非常适合于需要对缓存操作进行监控的场景，尤其是对于高性能、分布式应用程序，能有效地记录缓存的命中、写入和删除次数。



### 2 roundrobin/roundrobin.go

#### 2.1 代码详解

```go
package roundrobin

import (
	"fmt"
	"github.com/apex/log"
	"sync"
	"sync/atomic"
)

// RoundRobiner can pick a token from a list of tokens
type RoundRobiner interface {
	Pick() (*Token, error)
}

// New round robin implements with the given list of tokens
func New(tokens []string) RoundRobiner {
	log.Debugf("creating round robin with %d tokens", len(tokens))
	if len(tokens) == 0 {
		return &noTokenRoundRobin{}
	}
	result := make([]*Token, 0, len(tokens))
	for _, item := range tokens {
		result = append(result, NewToken(item))
	}
	return &realRoundRobin{tokens: result}
}

type realRoundRobin struct {
	tokens []*Token
	next   int64
}

func (rr *realRoundRobin) Pick() (*Token, error) {
	return rr.doPick(0)
}

func (rr *realRoundRobin) doPick(try int) (*Token, error) {
	if try > len(rr.tokens) {
		return nil, fmt.Errorf("no valid tokens left")
	}
	idx := atomic.LoadInt64(&rr.next)
	atomic.StoreInt64(&rr.next, (idx+1)%int64(len(rr.tokens)))
	if pick := rr.tokens[idx]; pick.OK() {
		log.Debugf("picked %s", pick.Key())
		return pick, nil
	}
	return rr.doPick(try + 1)
}

type noTokenRoundRobin struct {
}

func (rr *noTokenRoundRobin) Pick() (*Token, error) {
	return nil, nil
}

// Token is a github token
type Token struct {
	token string
	valid bool
	lock  sync.RWMutex
}

func NewToken(token string) *Token {
	return &Token{
		token: token,
		valid: true,
	}
}

// String returns the last 3 chars for the token
func (t *Token) String() string {
	return t.token[len(t.token)-3:]
}

// Key returns the actual token.
func (t *Token) Key() string {
	return t.token
}

// OK returns true if the token is valid
func (t *Token) OK() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.valid
}

// Invalidate invalidates the token.
func (t *Token) Invalidate() {
	log.Warnf("invalidate token '...%s'", t)
	t.lock.Lock()
	defer t.lock.Unlock()
	t.valid = false
}
```

这段代码实现了一个基于 GitHub Token 的 **圆形轮询（Round Robin）** 机制，具体包括以下几个主要部分：

- `RoundRobiner` 接口：定义了从 token 列表中选择一个有效 token 的方法 `Pick`。
- `realRoundRobin`：实际的圆形轮询实现，使用原子操作来保证对 `next` 索引的并发安全。
- `noTokenRoundRobin`：当没有 token 时的特殊处理，直接返回 `nil`。
- `Token` 类型：表示一个 token，包含有效性判断和失效功能。

**核心概念：**

- 圆形轮询：通过 `next` 索引，按照循环顺序轮询 token 列表。
- 原子操作：使用 `atomic.LoadInt64` 和 `atomic.StoreInt64` 来保证并发安全。
- 线程安全：通过 `sync.RWMutex` 实现对 token 的有效性字段的并发安全访问。

这个设计允许在多线程/并发环境下，安全高效地轮询有效的 token。在使用时，如果某个 token 无效，代码会自动跳过，直到找到有效的 token。

**其中需要注意的代码为 doPick 方法：**

```go
func (rr *realRoundRobin) doPick(try int) (*Token, error) {
    if try > len(rr.tokens) {
        return nil, fmt.Errorf("no valid tokens left")
    }
    idx := atomic.LoadInt64(&rr.next)
    atomic.StoreInt64(&rr.next, (idx+1)%int64(len(rr.tokens)))
    if pick := rr.tokens[idx]; pick.OK() {
        log.Debugf("picked %s", pick.Key())
        return pick, nil
    }
    return rr.doPick(try + 1)
}
```

**`doPick` 方法**：该方法尝试从 `tokens` 列表中选择一个有效的 `Token`。

- 它首先检查是否已经尝试选择了多次（`try > len(rr.tokens)`），如果超过最大尝试次数，则返回错误。
- 然后通过 `atomic.LoadInt64` 获取当前的 `next` 索引。
- 更新 `next` 索引，使用 `atomic.StoreInt64` 来保证索引更新是原子的。
- 如果当前选中的 `Token` (`rr.tokens[idx]`) 是有效的（`OK()` 方法返回 `true`），则返回该 `Token`。
- 如果选中的 `Token` 无效，则递归调用 `doPick`，增加 `try` 次数，尝试下一个 `Token`。

#### 2.2 单元测试

**之后，我们为 roundrobin.go 编写如下的单元测试**

```go
package roundrobin

import (
	"github.com/matryer/is"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	tokenA = "ghp_TokenA"
	tokenB = "ghp_TokenB"
	tokenC = "ghp_TokenC"
	tokenD = "ghp_TokenD"
)

var tokens = []string{tokenA, tokenB, tokenC, tokenD}

func TestRoundRobin(t *testing.T) {
	is := is.New(t)
	rr := New(tokens)
	a, b, c, d := exercise(t, rr, 100)

	for _, n := range []int64{a, b, c, d} {
		requireWithinRange(t, is, n, 23, 27)
	}
	is.Equal(int64(100), a+b+c+d)
}

func TestTokenString(t *testing.T) {
	is := is.New(t)
	is.Equal("enA", NewToken(tokenA).String())
	is.Equal("enB", NewToken(tokenB).String())
	is.Equal("enC", NewToken(tokenC).String())
	is.Equal("enD", NewToken(tokenD).String())
}

func TestNoTokens(t *testing.T) {
	is := is.New(t)
	rr := New([]string{})
	pick, err := rr.Pick()
	is.True(pick == nil) // pick should not nil
	is.NoErr(err)        // no error should be returned
}

func TestNoValidTokens(t *testing.T) {
	is := is.New(t)
	rr := New([]string{tokenA, tokenB})
	invalidateN(t, rr, 2)

	pick, err := rr.Pick()
	is.True(pick == nil) // pick should be nil
	is.True(err != nil)  // should err
}

func invalidateN(t *testing.T, rr RoundRobiner, n int) {
	t.Helper()
	is := is.New(t)
	for i := 0; i < n; i++ {
		pick, err := rr.Pick()
		is.True(pick != nil)
		is.NoErr(err)
		pick.Invalidate()
	}
}

func requireWithinRange(t *testing.T, is *is.I, n, low, high int64) {
	t.Helper()
	is.True(n >= low)  // n should be at least min
	is.True(n <= high) // n should be at most max
}

func exercise(t *testing.T, rr RoundRobiner, n int) (int64, int64, int64, int64) {
	t.Helper()
	is := is.New(t)

	var a, b, c, d int64
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			pick, err := rr.Pick()
			is.NoErr(err)
			is.True(pick != nil)
			switch pick.Key() {
			case tokenA:
				atomic.AddInt64(&a, 1)
			case tokenB:
				atomic.AddInt64(&b, 1)
			case tokenC:
				atomic.AddInt64(&c, 1)
			case tokenD:
				atomic.AddInt64(&d, 1)
			default:
				t.Error("Invalid pick:", pick)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return a, b, c, d
}
```

这段代码主要是为了测试一个 **轮询调度器**（`RoundRobin`）的功能，确保它在高并发情况下能够均衡地分配任务，且在没有 token 或所有 token 都无效时，能够正确处理异常情况。



### 3 github/github.go 

#### 3.1 代码详解 

```go
package github

import (
	"B1-StarCharts/config"
	"B1-StarCharts/internal/cache"
	"B1-StarCharts/internal/roundrobin"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
)

// ErrRateLimit happens when we rate limit github API.
var ErrRateLimit = errors.New("rate limit, please try again later")

// ErrGitHubAPI happens when fail to connect with github api
var ErrGitHubAPI = errors.New("failed to talk with github api")

// GitHub client struct.
type GitHub struct {
	tokens          roundrobin.RoundRobiner
	pageSize        int
	cache           *cache.Redis
	maxRageUsagePct int
}

var rateLimits = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "rate_limit_hits_total",
})

var effectiveEtags = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "effective_etag_uses_total",
})

var tokensCount = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "available_tokens",
})

var invalidatedTokens = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "invalidated_tokens_total",
})

var rateLimiter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "rate_limit_remaining",
}, []string{"token"})

func init() {
	prometheus.MustRegister(rateLimits, effectiveEtags, tokensCount, invalidatedTokens, rateLimiter)
}

// New github client
func New(config config.Config, cache *cache.Redis) *GitHub {
	tokensCount.Set(float64(len(config.GitHubTokens)))
	return &GitHub{
		tokens:          roundrobin.New(config.GitHubTokens),
		pageSize:        config.GitHubPageSize,
		cache:           cache,
		maxRageUsagePct: config.GitHubMaxRateUsagePct,
	}
}

const maxTries = 3

func (gh *GitHub) authorizedDo(req *http.Request, try int) (*http.Response, error) {
	if try > maxTries {
		return nil, fmt.Errorf("couldn't find a valid token")
	}
	token, err := gh.tokens.Pick()
	if err != nil || token == nil {
		log.WithError(err).Error("couldn't get a valid token")
		return http.DefaultClient.Do(req) // try unauthorized request
	}

	if err := gh.checkToken(token); err != nil {
		log.WithError(err).Error("couldn't check rate limit, try again")
		gh.authorizedDo(req, try+1)
	}

	// we got a valid token, use it: add into the req header
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (gh *GitHub) checkToken(token *roundrobin.Token) error {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/rate_limit", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token.Invalidate()
		invalidatedTokens.Inc()
		return fmt.Errorf("token is invalid")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	bst, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var limit rateLimit
	if err := json.Unmarshal(bst, &limit); err != nil {
		return err
	}
	rate := limit.Rate
	log.Debugf("%s rate %d/%d", token, rate.Remaining, rate.Limit)
	rateLimiter.WithLabelValues(token.String()).Set(float64(rate.Remaining))
	if isAboveTargetUsage(rate, gh.maxRageUsagePct) {
		return fmt.Errorf("token usage is too high %d/%d", rate.Remaining, rate.Limit)
	}
	return nil // allow at most x% rate limit usage
}

func isAboveTargetUsage(rate rate, target int) bool {
	return (rate.Remaining * 100 / rate.Limit) < target
}

type rateLimit struct {
	Rate rate `json:"rate"`
}

type rate struct {
	Remaining int `json:"remaining"`
	Limit     int `json:"limit"`
}
```

这段代码实现了一个 GitHub API 客户端 (`GitHub`)，使用轮询调度（`roundrobin`）来管理多个 GitHub 令牌（tokens）。它通过不同的 token 轮流进行请求，以避免单个 token 达到速率限制。同时，还集成了 Prometheus 来监控 API 调用的速率、令牌使用情况等。

下面我将逐步解释每个部分的功能：

##### 3.1.1 **GitHub 客户端结构 (`GitHub` struct)**

```go
type GitHub struct {
	tokens          roundrobin.RoundRobiner  // 轮询调度器，管理多个 GitHub 令牌
	pageSize        int                      // 每次请求返回的数据页大小
	cache           *cache.Redis             // 用于缓存的 Redis 客户端
	maxRageUsagePct int                      // 最大速率使用百分比限制
}
```

- `GitHub` 是客户端的核心结构，管理多个 GitHub API 令牌、请求的分页大小、缓存和速率限制等。
- `tokens` 是一个轮询调度器 (`RoundRobiner`)，用于在多个 GitHub 令牌之间轮流选择。
- `pageSize` 是 GitHub API 请求的分页大小，决定每次请求返回多少数据。
- `cache` 是一个 Redis 缓存，用于缓存数据（具体缓存操作不在这段代码中体现，但可能用于存储从 GitHub 获取的数据）。
- `maxRageUsagePct` 是一个限制，当 GitHub API 的剩余请求速率过低时，客户端会停止使用该令牌。

##### 3.1.2. **Prometheus 监控**

```go
var rateLimits = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "rate_limit_hits_total",
})
var effectiveEtags = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "effective_etag_uses_total",
})
var tokensCount = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "available_tokens",
})
var invalidatedTokens = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "invalidated_tokens_total",
})
var rateLimiter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "starcharts",
	Subsystem: "github",
	Name:      "rate_limit_remaining",
}, []string{"token"})
```

这些是 **Prometheus** 监控指标，用于跟踪 GitHub API 调用的各项数据：

- `rateLimits`：统计总的速率限制请求次数。
- `effectiveEtags`：统计有效的 ETag（缓存标记）使用次数。
- `tokensCount`：当前可用令牌的数量。
- `invalidatedTokens`：被标记为无效的令牌数量。
- `rateLimiter`：以令牌为标签，记录剩余的速率限制（剩余 API 调用次数）。

在 `init()` 函数中，这些指标被注册到 Prometheus 中，用于在后续监控中使用。

##### 3.1.3 **`New` 函数**

```go
func New(config config.Config, cache *cache.Redis) *GitHub {
	tokensCount.Set(float64(len(config.GitHubTokens)))
	return &GitHub{
		tokens:          roundrobin.New(config.GitHubTokens),
		pageSize:        config.GitHubPageSize,
		cache:           cache,
		maxRageUsagePct: config.GitHubMaxRateUsagePct,
	}
}
```

`New` 函数是 `GitHub` 客户端的构造函数。它初始化一个新的 `GitHub` 客户端，传入 GitHub 令牌、分页大小、Redis 缓存以及最大速率使用百分比。

- 它通过 `roundrobin.New(config.GitHubTokens)` 初始化了一个轮询调度器 `tokens`，这个调度器负责管理多个 GitHub 令牌，并在调用 API 时轮流使用它们。
- 它还将 `tokensCount` 设置为可用令牌的数量（来自配置）。

##### 3.1.4. **`authorizedDo` 函数**

```go
func (gh *GitHub) authorizedDo(req *http.Request, try int) (*http.Response, error) {
	if try > maxTries {
		return nil, fmt.Errorf("couldn't find a valid token")
	}
	token, err := gh.tokens.Pick()
	if err != nil || token == nil {
		log.WithError(err).Error("couldn't get a valid token")
		return http.DefaultClient.Do(req) // try unauthorized request
	}

	if err := gh.checkToken(token); err != nil {
		log.WithError(err).Error("couldn't check rate limit, try again")
		gh.authorizedDo(req, try+1)
	}

	// we got a valid token, use it: add into the req header
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}
```

- `authorizedDo` 是一个内部方法，负责执行带有 GitHub 令牌的请求。如果首次请求失败，它会使用轮询调度器重新尝试，直到成功或达到最大重试次数。
- 它首先尝试从 `tokens` 中选择一个可用的 token。
- 如果选中的 token 可用（未被标记为无效），它会在请求头中加入 `Authorization` 字段，并进行请求。
- 如果选中的 token 无效，它会再次尝试使用其他 token。
- 如果所有 token 都无效，或者超过最大重试次数（`maxTries`），则返回错误。

##### 3.1.5. **`checkToken` 函数**

```go
func (gh *GitHub) checkToken(token *roundrobin.Token) error {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/rate_limit", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("token %s", token.Key()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token.Invalidate()
		invalidatedTokens.Inc()
		return fmt.Errorf("token is invalid")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	bst, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var limit rateLimit
	if err := json.Unmarshal(bst, &limit); err != nil {
		return err
	}
	rate := limit.Rate
	log.Debugf("%s rate %d/%d", token, rate.Remaining, rate.Limit)
	rateLimiter.WithLabelValues(token.String()).Set(float64(rate.Remaining))
	if isAboveTargetUsage(rate, gh.maxRageUsagePct) {
		return fmt.Errorf("token usage is too high %d/%d", rate.Remaining, rate.Limit)
	}
	return nil // allow at most x% rate limit usage
}
```

- `checkToken` 检查指定的 token 是否有效，并获取其当前的速率限制。
- 它向 GitHub API 发送一个请求（`https://api.github.com/rate_limit`），并检查返回的速率限制数据。
- 如果返回的状态是 `Unauthorized`，表示该 token 无效，将其标记为无效。
- 如果剩余请求数超过了设定的阈值（`maxRageUsagePct`），则认为该 token 不可用。
- 如果状态是 `OK`，它会记录剩余的请求数，并将其保存到 Prometheus 中。

##### 3.1.6. **`isAboveTargetUsage` 函数**

```go
func isAboveTargetUsage(rate rate, target int) bool {
	return (rate.Remaining * 100 / rate.Limit) < target
}
```

- `isAboveTargetUsage` 用于判断当前 token 的剩余请求数是否超过了预设的使用百分比阈值（`target`）。
- 如果剩余请求数占总请求数的百分比大于 `target`，返回 `true`，表示该 token 的使用已达到上限。

##### 3.1.7. **`rateLimit` 和 `rate` 结构**

```go
type rateLimit struct {
	Rate rate `json:"rate"`
}

type rate struct {
	Remaining int `json:"remaining"`
	Limit     int `json:"limit"`
}
```

- 这些结构用于解析 GitHub API 返回的速率限制信息。`rate` 包含 `remaining`（剩余请求数）和 `limit`（总请求数限制）。



#### 3.2 单元测试

```go
package github

import (
	is "github.com/matryer/is"
	"testing"
)

func TestIsRateAboveLimit(t *testing.T) {
	is := is.New(t)

	is.Equal(false, isAboveTargetUsage(rate{
		Remaining: 4000,
		Limit:     5000,
	}, 50))

	...
}
```

该单元测试的目的是验证 `isAboveTargetUsage` 函数在处理 API 配额和目标百分比时的逻辑是否正确。

在上面测试样例中，期望 `isAboveTargetUsage` 返回 `false`，因为剩余配额占比高于目标值 50%。



### 4 github/repo.go 

#### 4.1 代码详解 

```go
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apex/log"
	"io"
	"net/http"
)

type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreateAt        string `json:"create_at"`
}

var ErrorNotFound = errors.New("repository not found")

// RepoDetails gets the given repository details.
func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {
	var repo Repository
	log := log.WithField("repo", name)

	var etag string
	etagKey := name + "_etag"

	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.WithError(err).Warnf("failed to get %s from cache", etagKey)
	}

	// etag can be nil or not nil
	resp, err := gh.makeRepoRequest(ctx, name, etag)
	if err != nil {
		return repo, err // return a nil repo
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return repo, err // return a nil repo
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		log.Info("not modified")
		effectiveEtags.Inc()
		err := gh.cache.Get(name, &repo)
		if err != nil {
			log.WithError(err).Warnf("failed to get %s from cache", name)
			// delete etag and try again
			if err := gh.cache.Delete(etagKey); err != nil {
				log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
			}
			return gh.RepoDetails(ctx, name)
		}
		// repo get from cache
		return repo, err
	case http.StatusForbidden:
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return repo, ErrRateLimit
	case http.StatusOK:
		if err := json.Unmarshal(bts, &repo); err != nil {
			return repo, err
		}
		if err := gh.cache.Put(name, repo); err != nil {
			log.WithError(err).Warnf("failed to cache %s", name)
		}
		// save a etag if it has etag
		etag = resp.Header.Get("etag")
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.WithError(err).Warnf("failed to cache %s", etagKey)
			}
		}
		return repo, nil
	case http.StatusNotFound:
		return repo, ErrorNotFound
	default:
		return repo, fmt.Errorf("%w : %v", ErrGitHubAPI, string(bts))
	}
}

func (gh *GitHub) makeRepoRequest(ctx context.Context, name, etag string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}
	return gh.authorizedDo(req, 0)
}
```

这段代码是一个实现与 GitHub API 交互的 Go 语言程序，目的是获取 GitHub 上某个仓库的详细信息（例如仓库的名称、星标数、创建时间等）。主要功能是通过缓存机制和 GitHub API 来提高效率，并处理 GitHub API 的不同响应状态。以下是详细的解析：

##### 4.1.1 **`Repository` 结构体**：

定义了仓库的基本信息：仓库的完整名称（`FullName`）、星标数量（`StargazersCount`）和创建时间（`CreateAt`）。

```go
type Repository struct {
    FullName        string `json:"full_name"`
    StargazersCount int    `json:"stargazers_count"`
    CreateAt        string `json:"create_at"`
}
```

##### 4.1.2 **`ErrorNotFound`**：

这是一个自定义的错误，表示找不到仓库（HTTP 404 错误时使用）。

```go
var ErrorNotFound = errors.New("repository not found")
```

##### 4.1.3 **`RepoDetails` 方法**：

该方法用于获取指定仓库（`name`）的详细信息。它首先会检查缓存中是否已有该仓库的相关数据，如果缓存中没有，则会向 GitHub API 发送请求以获取数据。如果 GitHub 返回的是 "未修改"（`304 Not Modified`）的状态，表示仓库数据没有改变，程序将尝试从缓存中读取数据。

```go
func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {
    var repo Repository
    log := log.WithField("repo", name)

    var etag string
    etagKey := name + "_etag"

    // 尝试从缓存中读取 ETag
    if err := gh.cache.Get(etagKey, &etag); err != nil {
        log.WithError(err).Warnf("failed to get %s from cache", etagKey)
    }

    // 向 GitHub API 发送请求
    resp, err := gh.makeRepoRequest(ctx, name, etag)
    if err != nil {
        return repo, err // 返回空的 Repository
    }

    bts, err := io.ReadAll(resp.Body)
    if err != nil {
        return repo, err // 返回空的 Repository
    }
    defer resp.Body.Close()

    // 根据 HTTP 响应状态进行不同的处理
    switch resp.StatusCode {
    case http.StatusNotModified:
        // 如果返回的状态是 304 Not Modified，表示仓库没有更新
        log.Info("not modified")
        effectiveEtags.Inc()
        err := gh.cache.Get(name, &repo)
        if err != nil {
            log.WithError(err).Warnf("failed to get %s from cache", name)
            // 删除 etag，并尝试再次获取仓库数据
            if err := gh.cache.Delete(etagKey); err != nil {
                log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
            }
            return gh.RepoDetails(ctx, name)
        }
        return repo, err
    case http.StatusForbidden:
        // 如果 GitHub 返回 403 Forbidden，表示达到了 API 的请求限制
        rateLimits.Inc()
        log.Warn("rate limit hit")
        return repo, ErrRateLimit
    case http.StatusOK:
        // 如果返回 200 OK，表示仓库数据已成功获取
        if err := json.Unmarshal(bts, &repo); err != nil {
            return repo, err
        }
        // 将仓库信息缓存
        if err := gh.cache.Put(name, repo); err != nil {
            log.WithError(err).Warnf("failed to cache %s", name)
        }
        // 保存 etag（如果有的话）
        etag = resp.Header.Get("etag")
        if etag != "" {
            if err := gh.cache.Put(etagKey, etag); err != nil {
                log.WithError(err).Warnf("failed to cache %s", etagKey)
            }
        }
        return repo, nil
    case http.StatusNotFound:
        // 如果返回 404 Not Found，表示仓库不存在
        return repo, ErrorNotFound
    default:
        // 其他情况，返回一个格式化的错误
        return repo, fmt.Errorf("%w : %v", ErrGitHubAPI, string(bts))
    }
}
```

**详细步骤：**

1. **从缓存读取 ETag**：
   - 通过 `gh.cache.Get` 方法尝试从缓存中获取与仓库名称相关的 `etag`（一个 GitHub API 用来标识资源状态的标识符）。如果缓存中没有该数据，会记录警告信息。
2. **向 GitHub 发送请求**：
   - 使用 `makeRepoRequest` 方法构造一个 HTTP 请求。此请求会附带一个 `If-None-Match` 头部字段，值为缓存中的 `etag`（如果存在）。如果 GitHub 上的仓库数据没有改变，API 会返回 HTTP 状态码 304（未修改），这样可以避免重复获取相同的数据。
3. **处理响应**：
   - 根据 GitHub API 返回的状态码，程序会执行不同的处理：
     - **HTTP 304（Not Modified）**：表示仓库数据未发生变化，程序从缓存中获取仓库数据。
     - **HTTP 403（Forbidden）**：表示已达到 API 请求限制，程序返回一个 `ErrRateLimit` 错误。
     - **HTTP 200（OK）**：表示成功获取到仓库数据，程序会将仓库信息解析为 `Repository` 类型，并缓存仓库数据和 `etag`。
     - **HTTP 404（Not Found）**：表示仓库不存在，程序返回一个 `ErrorNotFound` 错误。
     - **其他错误**：返回一个格式化的错误，包含 GitHub API 返回的响应体。
4. **缓存机制**：
   - 仓库信息和 `etag` 会被存储到缓存中，以便下次快速获取。如果 `etag` 没有变化，程序会直接从缓存中读取仓库信息，从而提高性能。

##### 4.1.4 ` makeRepoRequest` 方法

```go
func (gh *GitHub) makeRepoRequest(ctx context.Context, name, etag string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}
	return gh.authorizedDo(req, 0)
}
```

`makeRepoRequest` 方法的作用是：

1. 根据传入的仓库名称 `name` 构造 GitHub API 的请求 URL。
2. 创建一个 HTTP GET 请求，目标是获取该仓库的详细信息。
3. 如果提供了 `etag`，则将其添加到请求头中的 `If-None-Match` 字段，用于条件请求。
4. 最后，调用 `authorizedDo` 方法发送请求并返回响应。

##### 4.1.5 关键点

- **缓存**：利用缓存机制（`gh.cache.Get`、`gh.cache.Put`）避免多次请求同一个仓库，提升效率。
- **ETag**：使用 `etag` 来判断仓库数据是否已经更新，避免不必要的重复请求。
- **错误处理**：根据不同的 HTTP 状态码（如 404、403、304）做不同的错误处理和日志记录。
- **递归请求**：在某些错误情况下（如缓存失效），可能会重新尝试请求仓库的详细信息。



#### 4.2 单元测试

```go
package github

import (
	"B1-StarCharts/config"
	"B1-StarCharts/internal/cache"
	"context"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/matryer/is"
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func TestRepoDetails(t *testing.T) {
	defer gock.Off()

	repo := Repository{
		FullName:        "test/test",
		CreateAt:        "2008-02-28T20:40:04Z",
		StargazersCount: 3811,
	}

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()

	gt := New(config, cache)
	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	t.Run("get repo details from api", func(t *testing.T) {
		is := is.New(t)
		gock.New("https://api.github.com").
			Get("/repos/test/test").
			Reply(200).
			JSON(repo)
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		is.NoErr(err) // should not fail to get from api
	})
}

func TestRepoDetails_APIfailure(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	gock.New("https://api.github.com").
		Get("/repos/test/test").
		Reply(404)

	gock.New("https://api.github.com").
		Get("/repos/private/private").
		Reply(403)

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()

	gt := New(config, cache)

	t.Run("set error if api return 404", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.RepoDetails(context.TODO(), "test/test")
		is.True(err != nil) // Expected error
	})

	t.Run("set error if api return 403", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.RepoDetails(context.TODO(), "private/private")
		is.True(err != nil) // Expected error
	})
}

func TestRepoDetails_WithAuthToken(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	repo := Repository{
		FullName:        "aasm/aasm",
		CreateAt:        "2008-02-28T20:40:04Z",
		StargazersCount: 3811,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/private").
		Reply(200).
		JSON(repo)

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()

	gt := New(config, cache)
	gt.tokens = roundrobin.New([]string{"12345"})
  
	t.Run("get repo with auth token", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.RepoDetails(context.TODO(), "test/private")
		is.NoErr(err) // should not fail to get from api with auth token
	})
}
```

这段代码包含了三个单元测试，它们分别测试了 GitHub 仓库信息获取方法（`RepoDetails`）在不同场景下的表现。以下是对每个测试用例的简洁总结：

1. **`TestRepoDetails`**:

- **测试目标**：验证正常情况下从 GitHub API 获取仓库详情的功能。
- **测试内容**：模拟了 GitHub API 返回一个成功的仓库响应（`200 OK`），并检查 `RepoDetails` 方法是否正确获取仓库信息，且没有报错。
- **预期结果**：API 请求成功，`RepoDetails` 应该能够正常返回仓库信息且不出现错误。

2. **`TestRepoDetails_APIfailure`**:

- **测试目标**：验证 GitHub API 返回错误响应时的处理情况。
- **测试内容** ：
  - 模拟 GitHub API 返回 `404 Not Found` 错误，检查 `RepoDetails` 是否正确处理该错误。
  - 模拟 GitHub API 返回 `403 Forbidden` 错误，检查 `RepoDetails` 是否正确处理该错误。
- **预期结果**：当 API 返回 404 或 403 错误时，`RepoDetails` 应该正确返回错误。

3. **`TestRepoDetails_WithAuthToken`**:

- **测试目标**：验证使用认证令牌（Auth Token）时从 GitHub API 获取仓库详情的功能。
- **测试内容**：模拟了 GitHub API 返回一个成功的仓库响应（`200 OK`），并检查在使用认证令牌的情况下，`RepoDetails` 方法是否能够正确获取仓库信息。
- **预期结果**：即使使用认证令牌，`RepoDetails` 也应正常工作且不出现错误。

总结：

- **`TestRepoDetails`**：测试正常的 API 请求。
- **`TestRepoDetails_APIfailure`**：测试 API 返回错误时的错误处理。
- **`TestRepoDetails_WithAuthToken`**：测试带认证令牌时的请求行为。



#### 4.3 304 Not Modified 作用

**ETag HTTP 响应头是资源的特定版本的标识符。这可以让缓存更高效，并节省带宽，因为如果内容没有改变，Web 服务器不需要发送完整的响应。**

+ `req.Header.Add("If-None-Match", etag)` 这个 HTTP 请求头的作用是启用条件请求，确保只有当资源的 **ETag** 值与提供的 **ETag** 不匹配时，服务器才会返回资源内容。
+ 如果 ETag 相匹配，表示资源没有更改，服务器会返回 304 Not Modified 响应，告知客户端资源没有变化，可以继续使用缓存的版本。

**详细解释：**

- **ETag**（实体标签）是服务器为某个资源生成的唯一标识符，通常是资源内容的哈希值。当资源发生变化时，ETag 也会发生变化。
- **If-None-Match** 是一个 HTTP 请求头，用来向服务器传递一个或多个 ETag 值。服务器会将传递的 ETag 与资源的当前 ETag 值进行比对：
  - 如果传递的 ETag 和当前资源的 ETag **匹配**，服务器会返回 `304 Not Modified` 响应，表示资源未发生变化，客户端可以继续使用本地缓存的副本。
  - 如果传递的 ETag 和资源的 ETag **不匹配**，服务器会返回完整的资源内容和 200 OK 状态码。

**典型使用场景：**

- **缓存优化**：通过这种机制，客户端可以避免重复下载未改变的资源，从而减少带宽消耗，提升响应速度。
- **减少无谓的资源加载**：比如在实现增量更新或数据同步时，可以利用 ETag 和 `If-None-Match` 来避免客户端重新下载已缓存的资源。

示例：假设一个客户端请求一个资源，服务器返回一个带有 ETag 的响应：

```
HTTP/1.1 200 OK
ETag: "12345"
Content-Type: application/json
Content-Length: 200
```

客户端随后再次请求该资源，但这次请求会加上 `If-None-Match` 头：

```
GET /resource HTTP/1.1
If-None-Match: "12345"
```

服务器收到请求后，如果资源的 ETag 仍然是 `"12345"`，它会返回：

```
HTTP/1.1 304 Not Modified
```

如果资源的 ETag 已变，则会返回新的资源内容和状态码 200 OK。

总结：`If-None-Match` 是一种缓存控制机制，帮助客户端通过 ETag 值判断资源是否发生变化，减少不必要的资源传输。



### 5 github/star.go 

#### 5.1 代码详解 

```go
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/apex/log"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

var (
	errNoMorePages  = errors.New("no more pages to get")
	ErrTooManyStars = errors.New("repo has too many stargazers, github won't allow us to list all stars")
)

// Stargazer is a star at a given time.
type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
}

// Stargazers returns all the stargazers of a given repo.
func (gh *GitHub) Stargazers(ctx context.Context, repo Repository) (stars []Stargazer, err error) {
	if gh.totalPages(repo) > 400 {
		return stars, ErrTooManyStars
	}

	var (
		wg   errgroup.Group
		lock sync.Mutex
	)

	wg.SetLimit(4)
	for page := 1; page <= gh.lastPage(repo); page++ {
		page := page
		wg.Go(func() error {
			result, err := gh.getStargazersPage(ctx, repo, page)
			if errors.Is(err, errNoMorePages) {
				return nil
			}
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			stars = append(stars, result...)
			return nil
		})
	}
	err = wg.Wait()
	sort.Slice(stars, func(i, j int) bool {
		return stars[i].StarredAt.Before(stars[j].StarredAt)
	})
	return
}

// - get last modified from cache
//   - if exists, hit api with it
//     - if it returns 304, get from cache
//       - if succeeds, return it
//       - if fails, it means we dont have that page in cache, hit api again
//         - if succeeds, cache and return both the api and header
//         - if fails, return error
//   - if not exists, hit api
//     - if succeeds, cache and return both the api and header
//     - if fails, return error

// nolint: funlen
// TODO: refactor.
func (gh *GitHub) getStargazersPage(ctx context.Context, repo Repository, page int) ([]Stargazer, error) {
	log := log.WithField("repo", repo.FullName).WithField("page", page)
	defer log.Trace("get page").Stop(nil)

	var stars []Stargazer
	key := fmt.Sprintf("%s_%d", repo.FullName, page)
	etagKey := fmt.Sprintf("%s_%d", repo.FullName, page) + "_etag"

	var etag string
	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.WithError(err).Warnf("failed to get %s from cache", etagKey)
	}

	resp, err := gh.makeStarPageRequest(ctx, repo, page, etag)
	if err != nil {
		return stars, err
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		effectiveEtags.Inc()
		log.Info("not modified")
		// try to get stars from cache directly
		err := gh.cache.Get(key, &stars)
		if err != nil {
			log.WithError(err).Warnf("failed to get %s from cache", key)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
			}
			return gh.getStargazersPage(ctx, repo, page)
		}
		return stars, err
	case http.StatusForbidden:
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return stars, ErrRateLimit
	case http.StatusOK:
		if err := json.Unmarshal(bts, &stars); err != nil {
			return stars, err
		}
		if len(stars) == 0 {
			return stars, errNoMorePages
		}
		if err := gh.cache.Put(key, stars); err != nil {
			log.WithError(err).Warnf("failed to cache %s", key)
		}

		etag = resp.Header.Get("etag")
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.WithError(err).Warnf("failed to cache %s", etagKey)
			}
		}
		return stars, nil
	default:
		return stars, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}

}

func (gh *GitHub) totalPages(repo Repository) int {
	return repo.StargazersCount / gh.pageSize
}

func (gh *GitHub) lastPage(repo Repository) int {
	return gh.totalPages(repo) + 1
}

func (gh *GitHub) makeStarPageRequest(ctx context.Context, repo Repository, page int, etag string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
		repo.FullName,
		page,
		gh.pageSize,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}
	return gh.authorizedDo(req, 0)
}
```

这段代码主要用于从 GitHub API 获取指定仓库的所有 Stargazers（关注者/点赞者）。GitHub 对于大规模仓库的 Stargazers 数据进行分页限制，因此需要分多次请求来获取所有的 Stargazers 信息。同时，代码也实现了缓存机制，避免重复请求 GitHub API，提高效率。

以下是对这段代码的详细解释，这段代码支持分页、缓存、条件请求和并发处理：

1. **Stargazers**：这个方法用于获取指定 GitHub 仓库的所有 Stargazers 信息（即关注该仓库的用户列表）。通过分页方式逐步获取。
2. **并发请求控制**：通过 `errgroup.Group` 限制并发请求的数量，最多同时发起 4 个请求，这有助于提高请求效率并避免超过 GitHub API 的速率限制。
3. **缓存**：通过缓存来存储每一页的 Stargazers 数据和 `etag`，避免重复请求，从而提高效率。

##### 5.1.1. **`Stargazers`**：

```go
// Stargazers returns all the stargazers of a given repo.
func (gh *GitHub) Stargazers(ctx context.Context, repo Repository) (stars []Stargazer, err error) {
	if gh.totalPages(repo) > 400 {
		return stars, ErrTooManyStars
	}

	var (
		wg   errgroup.Group
		lock sync.Mutex
	)

	wg.SetLimit(4)
	for page := 1; page <= gh.lastPage(repo); page++ {
		page := page
		wg.Go(func() error {
			result, err := gh.getStargazersPage(ctx, repo, page)
			if errors.Is(err, errNoMorePages) {
				return nil
			}
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			stars = append(stars, result...)
			return nil
		})
	}
	err = wg.Wait()
	sort.Slice(stars, func(i, j int) bool {
		return stars[i].StarredAt.Before(stars[j].StarredAt)
	})
	return
}
```

这个方法获取所有的 Stargazers 信息，做了以下几件事：

- **检查仓库的 Stargazer 数量**：如果仓库的 Stargazer 数量超过 400 页，则直接返回 `ErrTooManyStars` 错误。
- **并发分页请求**：通过循环分页请求所有的 Stargazers（每页的数据量通过 `gh.pageSize` 控制）。它通过 `errgroup.Group` 来实现并发请求，每次最多同时发起 4 个请求。
- **合并结果**：当请求的每一页数据返回后，它将结果存储在 `stars` 列表中，并在所有请求完成后对 `stars` 进行排序（按照 `StarredAt` 时间排序）。
- **返回排序后的 Stargazers 数据**：所有分页请求完成后，返回合并并排序后的 Stargazers 列表。

##### 5.1.2. **`getStargazersPage`**：

```go
func (gh *GitHub) getStargazersPage(ctx context.Context, repo Repository, page int) ([]Stargazer, error) {
	log := log.WithField("repo", repo.FullName).WithField("page", page)
	defer log.Trace("get page").Stop(nil)

	var stars []Stargazer
	key := fmt.Sprintf("%s_%d", repo.FullName, page)
	etagKey := fmt.Sprintf("%s_%d", repo.FullName, page) + "_etag"

	var etag string
	if err := gh.cache.Get(etagKey, &etag); err != nil {
		log.WithError(err).Warnf("failed to get %s from cache", etagKey)
	}

	resp, err := gh.makeStarPageRequest(ctx, repo, page, etag)
	if err != nil {
		return stars, err
	}

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		effectiveEtags.Inc()
		log.Info("not modified")
		// try to get stars from cache directly
		err := gh.cache.Get(key, &stars)
		if err != nil {
			log.WithError(err).Warnf("failed to get %s from cache", key)
			if err := gh.cache.Delete(etagKey); err != nil {
				log.WithError(err).Warnf("failed to delete %s from cache", etagKey)
			}
			return gh.getStargazersPage(ctx, repo, page)
		}
		return stars, err
	case http.StatusForbidden:
		rateLimits.Inc()
		log.Warn("rate limit hit")
		return stars, ErrRateLimit
	case http.StatusOK:
		if err := json.Unmarshal(bts, &stars); err != nil {
			return stars, err
		}
		if len(stars) == 0 {
			return stars, errNoMorePages
		}
		if err := gh.cache.Put(key, stars); err != nil {
			log.WithError(err).Warnf("failed to cache %s", key)
		}

		etag = resp.Header.Get("etag")
		if etag != "" {
			if err := gh.cache.Put(etagKey, etag); err != nil {
				log.WithError(err).Warnf("failed to cache %s", etagKey)
			}
		}
		return stars, nil
	default:
		return stars, fmt.Errorf("%w: %v", ErrGitHubAPI, string(bts))
	}
}
```

这个方法用于获取某一页的 Stargazers 数据。它涉及缓存、条件请求和 API 请求。具体流程如下：

- **缓存判断**：首先，它检查是否可以从缓存中获取页面的 `etag` 和数据。如果可以，则利用这些信息进行条件请求，避免重复请求 GitHub API。
- **条件请求**：如果 `etag` 存在，方法会添加 `If-None-Match` 请求头来告知 GitHub 仅在数据发生变化时返回新的数据。如果 GitHub 返回 `304 Not Modified`（即数据未改变），则直接从缓存获取数据。
- **API 请求**：如果缓存中没有数据或 `etag`，方法会发送 HTTP 请求获取指定页面的 Stargazers 数据。
- **缓存更新**：如果成功从 API 获取数据，则将结果和 `etag` 存储到缓存中，供下次使用。
- **错误处理**：如果 API 返回 404 或其他错误，方法会处理并返回相应错误。

##### 5.1.3. **`makeStarPageRequest`**：

```go
func (gh *GitHub) makeStarPageRequest(ctx context.Context, repo Repository, page int, etag string) (*http.Response, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
		repo.FullName,
		page,
		gh.pageSize,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if etag != "" {
		req.Header.Add("If-None-Match", etag)
	}
	return gh.authorizedDo(req, 0)
}
```

该方法构建并发送请求来获取指定页面的 Stargazers 数据：

- **请求 URL 构造**：使用传入的 `repo` 名称和 `page` 参数构建 GitHub API 请求的 URL。请求 URL 中包括 `page` 和每页数据量 `per_page`（由 `gh.pageSize` 控制）。
- **设置请求头**：请求头中设置 `Accept` 字段为 `application/vnd.github.v3.star+json`，指示 GitHub 返回 Stargazers 的专用格式。如果 `etag` 存在，则在请求中添加 `If-None-Match` 头来使用条件请求。
- **发送请求**：调用 `authorizedDo` 方法发送请求，并返回响应。

##### 5.1.4. **分页计算**：

```go
func (gh *GitHub) totalPages(repo Repository) int {
	return repo.StargazersCount / gh.pageSize
}

func (gh *GitHub) lastPage(repo Repository) int {
	return gh.totalPages(repo) + 1
}
```

- **`totalPages`**：计算仓库的 Stargazers 分页数，公式是将总的 Stargazers 数量（`repo.StargazersCount`）除以每页的数据量（`gh.pageSize`）。
- **`lastPage`**：返回最后一页的页码，等于总页数加 1。

##### 5.1.5 错误处理：

- **`errNoMorePages`**：当没有更多分页数据时返回此错误。
- **`ErrTooManyStars`**：如果仓库的 Stargazers 数量超过 400 页，则返回此错误，表示无法获取所有 Stargazers。
- **`ErrRateLimit`**：如果请求由于 GitHub API 的速率限制而被拒绝，返回此错误。
- **`ErrGitHubAPI`**：用于标记从 GitHub API 返回的其他错误。



#### 5.2 单元测试

```go
package github

import (
	"B1-StarCharts/config"
	"B1-StarCharts/internal/cache"
	"B1-StarCharts/internal/roundrobin"
	"context"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/matryer/is"
	"gopkg.in/h2non/gock.v1"
	"testing"
	"time"
)

func TestStargazers(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	stargazers := []Stargazer{
		{StarredAt: time.Now()},
		{StarredAt: time.Now()},
	}

	repo := Repository{
		FullName:        "test/test",
		CreateAt:        "2008-02-28T20:40:04Z",
		StargazersCount: 2,
	}

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()

	gt := New(config, cache)
	t.Run("get stargazers from api", func(t *testing.T) {
		is := is.New(t)
		gock.New("https://api.github.com").
			Get("/repos/test/test/stargazers").
			Reply(200).
			JSON(stargazers)
		_, err := gt.Stargazers(context.TODO(), repo)
		is.NoErr(err)
	})

	t.Run("get stargazers from cache", func(t *testing.T) {
		is := is.New(t)
		is.NoErr(cache.Put(repo.FullName+"_1_etag", "asdasd"))
		gock.New("https://api.github.com").
			Get("/repos/test/test/stargazers").
			MatchHeader("If-None-Match", "asdasd").
			Reply(304).
			JSON([]Stargazer{})
		_, err := gt.Stargazers(context.TODO(), repo)
		is.NoErr(err)
	})
}

func TestStargazers_EmptyResponseOnPagination(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4001}})

	stargazers := []Stargazer{
		{StarredAt: time.Now()},
		{StarredAt: time.Now()},
	}

	repo := Repository{
		FullName:        "test/test",
		CreateAt:        "2008-02-28T20:40:04Z",
		StargazersCount: 3,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		MatchParam("page", "1").
		MatchParam("per_page", "2").
		Reply(200).
		JSON(stargazers)

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		MatchParam("page", "2").
		MatchParam("per_page", "2").
		Reply(200).
		JSON([]Stargazer{})

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)

	// set some features for the obj - github(gt)
	gt.pageSize = 2
	gt.tokens = roundrobin.New([]string{"12345"})
	t.Run("get stargazers from api", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.Stargazers(context.TODO(), repo)
		is.NoErr(err) // should not have errored
	})
}

func TestStargazers_APIFailure(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		JSON(rateLimit{rate{Limit: 5000, Remaining: 4000}})

	repo1 := Repository{
		FullName:        "test/test",
		CreateAt:        "2008-02-28T20:40:04Z",
		StargazersCount: 3,
	}

	repo2 := Repository{
		FullName:        "private/private",
		CreateAt:        "2008-02-28T20:40:04Z",
		StargazersCount: 3,
	}

	gock.New("https://api.github.com").
		Get("/repos/test/test/stargazers").
		Persist().
		Reply(404).
		JSON([]Stargazer{})

	gock.New("https://api.github.com").
		Get("/repos/private/private/stargazers").
		Persist().
		Reply(403).
		JSON([]Stargazer{})

	mr, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	config := config.Get()
	cache := cache.New(rc)
	defer cache.Close()
	gt := New(config, cache)

	t.Run("set error if api return 404", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.Stargazers(context.TODO(), repo1)
		is.True(err != nil) // should not have errored
	})

	t.Run("set error if api return 403", func(t *testing.T) {
		is := is.New(t)
		_, err := gt.Stargazers(context.TODO(), repo2)
		is.True(err != nil) // should not have errored
	})
}
```

1. `TestStargazers`— **从 API 获取 Stargazers 并使用缓存**

- 目的：测试 `Stargazers` 方法从 GitHub API 获取 Stargazers 数据的正常流程，并且在缓存中获取数据时，能正确处理缓存。
- 测试内容：
  - **从 API 获取 Stargazers**：模拟 GitHub 返回的一个正常的 `200 OK` 响应，包含一个包含 Stargazers 的 JSON 数据。测试 `Stargazers` 方法是否正确从 API 获取并返回 Stargazers。
  - **从缓存获取 Stargazers**：模拟缓存中已经有了某个仓库的 `etag`，并且 GitHub API 返回 `304 Not Modified`，此时 `Stargazers` 方法应该从缓存中获取数据，而不是重新请求 API。
- 预期结果：两种情况都应该成功，无错误。

2. `TestStargazers_EmptyResponseOnPagination` — **分页请求并处理空的分页结果**

- 目的：测试在分页请求的过程中，如果某一页没有返回 Stargazers（即空响应），方法是否能正确处理。
- 测试内容：
  - **API 返回分页数据**：模拟了第一个分页返回正常数据，第二个分页返回空数据。通过分页请求获取 Stargazers，检查是否能正确处理分页逻辑。
  - **GitHub API 分页限制**：设置 `gt.pageSize = 2`，每页请求 2 个 Stargazers，第二页请求时返回空的 Stargazers 列表。
- 预期结果：`Stargazers` 方法应该能够正确处理空的分页结果并正常返回，且不发生错误。

3. `TestStargazers_APIFailure` — **处理 API 错误（404 和 403 错误）**

- 目的：测试当 GitHub API 返回错误（如 404 或 403 错误）时，`Stargazers` 方法如何处理这些错误。
- 测试内容：
  - **API 返回 404 错误**：模拟 GitHub 返回 `404 Not Found` 错误，检查 `Stargazers` 方法是否能正确处理该错误并返回相应的错误信息。
  - **API 返回 403 错误**：模拟 GitHub 返回 `403 Forbidden` 错误，检查 `Stargazers` 方法是否能正确处理该错误并返回相应的错误信息。
- 预期结果：当 GitHub 返回 404 或 403 错误时，`Stargazers` 方法应该返回适当的错误，而不会继续请求或获取数据。








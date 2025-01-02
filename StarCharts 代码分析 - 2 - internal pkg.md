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
	return (rate.Remaining * 100 / rate.Limit) > target
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
	return (rate.Remaining * 100 / rate.Limit) > target
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



### 4 github/repo.go 

#### 4.1 代码详解 

#### 4.2 单元测试



### 5 github/star.go 

#### 5.1 代码详解 

#### 5.2 单元测试
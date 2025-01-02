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

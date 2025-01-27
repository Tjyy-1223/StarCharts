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

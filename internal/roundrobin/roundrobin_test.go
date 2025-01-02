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

package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// Per limit n events per duration.
func Per(n int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(n))
}

type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

// APIConnection shows how a multi-dimension rate limiter works.
type APIConnection struct {
	networkLimiter,
	diskLimiter,
	apiLimiter RateLimiter
}

func Open() *APIConnection {
	return &APIConnection{
		apiLimiter: MultiLimiter(
			rate.NewLimiter(Per(2, time.Second), 2),
			rate.NewLimiter(Per(10, time.Minute), 10),
		),
		diskLimiter: MultiLimiter(
			rate.NewLimiter(rate.Limit(1), 1),
		),
		networkLimiter: MultiLimiter(
			rate.NewLimiter(Per(3, time.Second), 3),
		),
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	// request to ReadFile is limited by api and disk rate limiters.
	err := MultiLimiter(a.apiLimiter, a.diskLimiter).Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	// request to ResolveAddress is limited by api and net rate limiters.
	err := MultiLimiter(a.apiLimiter, a.networkLimiter).Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func TestMultiDimensionRateLimiting(_ *testing.T) {
	defer log.Printf("Done.")

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	api := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := api.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
			log.Printf("Readfile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := api.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
}
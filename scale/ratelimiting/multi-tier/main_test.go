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

// multiLimiter shows how to implement a multi-tier rate limiting. To do this,
// it's easier to keep the limiters separate and then combine them into one
// rate limiter that manages the interaction for us. So we created multiLimiter
// as a simple aggregate rate limiter.
//
// multiLimiter satisfy and aggregate RateLimiter interface, enable it to
// aggregate other multiLimiter recursively. (Composition pattern @??).
type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		// Since we've sorted the limiters by rate, request will be restricted
		// by the first strict-enough rate limiter.
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Limit return the most restrictive limit.
func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

// MultiLimiter aggregate multiple RateLimiter into one.
func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	// Sort by rate, stricter first.
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

// APIConnection shows how a multi-tier rate limiter works.
type APIConnection struct {
	rateLimiter RateLimiter
}

func Open() *APIConnection {
	secondLimiter := rate.NewLimiter(Per(2, time.Second), 1)	// 2 event per sec
	minuteLimiter := rate.NewLimiter(Per(10, time.Minute), 10)	// 10 event per minute
	return &APIConnection{
		rateLimiter: MultiLimiter(secondLimiter, minuteLimiter),
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

func TestMultiTierRateLimiting(_ *testing.T) {
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
package main

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"golang.org/x/time/rate"
)

// APIConnection shows how a one-tier rate limiter works. It use x/time/rate
// package which provide a token bucket rate limiter.
//
// To understand this example, first read x/time/rate API documentation. See:
//
// 		go doc x/time/rate
//
// Normally a rate limiter would be running on a server so the users couldn’t
// trivially bypass it. Production systems might also include a client-side
// rate limiter to help prevent the client from making unnecessary calls only
// to be denied, but that is an optimization.
//
// For demonstration, a client-side rate limiter keeps things simple.
type APIConnection struct {
	rateLimiter *rate.Limiter
}

func Open() *APIConnection {
	return &APIConnection{
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1),	// one event per second.
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

func TestOneTierRateLimiting(_ *testing.T) {
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

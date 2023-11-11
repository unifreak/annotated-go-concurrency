package main

import (
	"context"
	"fmt"
	"testing"
)

// TestDataBag shows the basic usage of context request-scoped data.
//
// Becuase context's Key and Value are both defined as interface{}, and we don't
// use a customized type for Key here, we lose Go's type safety when retrieving
// values. See typedkey/.
func TestDataBag(_ *testing.T) {
	ProcessRequest("jane", "abc123")
}

func ProcessRequest(userID, authToken string) {
	ctx := context.WithValue(context.Background(), "userID", userID)
	ctx = context.WithValue(ctx, "authToken", authToken)
	HandelResponse(ctx)
}

func HandelResponse(ctx context.Context) {
	fmt.Printf(
		"handling response for %v (%v)\n",
		ctx.Value("userID"),
		ctx.Value("authToken"),
	)
}


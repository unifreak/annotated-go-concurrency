package main

import (
	"context"
	"fmt"
	"testing"
)

type ctxKey int // unexported Key type, to avoid conflict with other packages.

const (
	ctxUserID    ctxKey = iota // key for storing UserID
	ctxAuthToken               // key for storing AuthToken
)

// Exported Value accessors for our Key types.
func UserID(c context.Context) string {
	return c.Value(ctxUserID).(string)
}

func AuthToken(c context.Context) string {
	return c.Value(ctxAuthToken).(string)
}

func ProcessRequest(userID, authToken string) {
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxAuthToken, authToken)
	HandleResponse(ctx)
}

func HandleResponse(ctx context.Context) {
	fmt.Printf(
	 	"handling response for %v (auth: %v)\n",
		UserID(ctx),
		AuthToken(ctx),
	)
}

// TestCustomTypedKey show how to safe-guard our keys with customized type.
//
// It has the following problem:
//
// Let’s say HandleResponse did live in another package named response, and
// let’s say the package ProcessRequest lived in a package named process. The
// process package would have to import the response package to make the call
// to HandleResponse, but HandleResponse would have no way to access the
// accessor functions defined in the process package because importing process
// would form a **circular dependency**.
//
// This problem will coerces the architecture into creating packages centered
// around data types that are imported from multiple locations. This certainly
// isn’t a bad thing, but it’s something to be aware of.
func TestCustomTypedKey(_ *testing.T) {
	ProcessRequest("jane", "abc123")
}

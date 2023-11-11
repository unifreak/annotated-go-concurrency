// Pool is also useful is for warming a cache of pre-allocated objects for
// operations that must run as quickly as possible. In this case, instead of
// trying to guard the host machine’s memory by constraining the number of
// objects created, we’re trying to guard consumers’ time by front-loading the
// time it takes to get a reference to another object.

// run: go test -benchtime=10s -bench=.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"
	"net"
)

// connectToService emulate a time comsuming connection to backend service, like
// DB.
func connectToService() interface{} {
	time.Sleep(1 * time.Second)
	return struct{}{}
}

func init() {
	coldDaemon := coldStartNetworkDaemon()
	coldDaemon.Wait()

	warmDaemon := warmStartNetworkDaemon()
	warmDaemon.Wait()
}

// ColdStart ---------------------------------------------------------

func coldStartNetworkDaemon() *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)	// for simplicity of benchmark, only allow one connection at a time
	go func() {
		server, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Fatalf("cannot listen: %v", err)
		}
		defer server.Close()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %v", err)
				continue
			}
			connectToService()
			fmt.Fprintln(conn, "")
			conn.Close()
		}
	}()
	return &wg
}


func BenchmarkColdStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			b.Fatalf("cannot dail host: %v", err)
		}
		if _, err := ioutil.ReadAll(conn); err != nil {
			b.Fatalf("cannot read: %v", err)
		}
		conn.Close()
	}
}

// WarmStart ---------------------------------------------------------

func warmServiceConnCache() *sync.Pool {
	p := &sync.Pool {	// backend service connection pool
		New: connectToService,
	}
	for i := 0; i < 10; i++ {
		p.Put(p.New())
	}
	return p
}

func warmStartNetworkDaemon() *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		connPool := warmServiceConnCache()

		server, err := net.Listen("tcp", "localhost:8181")
		if err != nil {
			log.Fatalf("caonnot listen: %v", err)
		}
		defer server.Close()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %v", err)
				continue
			}
			svcConn := connPool.Get()
			fmt.Fprintln(conn, "")
			connPool.Put(svcConn)
			conn.Close()
		}
	}()
	return &wg
}

func BenchmarkWarmStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", "localhost:8181")
		if err != nil {
			b.Fatalf("cannot dail host: %v", err)
		}
		if _, err := ioutil.ReadAll(conn); err != nil {
			b.Fatalf("cannot read: %v", err)
		}
		conn.Close()
	}
}
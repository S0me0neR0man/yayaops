package main

import (
	"context"
	"github.com/S0me0neR0man/yayaops/internal/client"
	"time"
)

// main
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := client.New().Start(ctx)
	time.Sleep(40 * time.Minute)
	cancel()
	c.WaitShutdown()
}

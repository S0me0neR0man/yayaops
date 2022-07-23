package main

import (
	"context"
	"github.com/S0me0neR0man/yayaops/internal/client"
)

// main
func main() {
	ctx, _ /*cancel*/ := context.WithCancel(context.Background())
	_ = client.New().Start(ctx)
	//time.Sleep(5 * time.Second)
	//cancel()
	//c.WaitShutdown()
}

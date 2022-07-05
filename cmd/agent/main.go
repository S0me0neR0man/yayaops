package main

import (
	"context"
	"github.com/S0me0neR0man/yayaops/internal/client"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	client.GetEngine().Start(ctx)
	time.Sleep(15 * time.Second)
	cancel()
}

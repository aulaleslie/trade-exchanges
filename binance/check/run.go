package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aulaleslie/trade-exchanges/binance"
	"go.uber.org/zap"
)

func main2() {
	log.Print("Starting...")
	client := binance.NewBinanceLong(binance.OriginalBinanceURLs, "", "", zap.NewExample())
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.WatchSymbolPrice(
		ctx,
		"BTCUSDT",
	)
	if err != nil {
		log.Printf("Open price watcher err: %v", err)
	}

	cancelTimer := time.NewTimer(time.Minute / 2)
	timer := time.NewTimer(time.Minute)

	streamClosed := false
	for {
		select {
		case ev, ok := <-stream:
			if ok {
				log.Printf("Event = %v", ev)
			} else if !streamClosed {
				log.Print("Stream closed")
				streamClosed = true
			}
		case <-cancelTimer.C:
			log.Print("Cancelling")
			cancel()
		case <-timer.C:
			log.Print("Exit")
			return
		}
	}
}

func main() {
	log.Print("Starting...")
	apiKey, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	client := binance.NewBinanceLong(binance.OriginalBinanceURLs, apiKey, secret, zap.NewExample())
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.WatchOrdersStatuses(ctx)
	if err != nil {
		log.Printf("Open orders watcher err: %v", err)
	}

	cancelTimer := time.NewTimer(10 * time.Second)
	timer := time.NewTimer(40 * time.Second)

	streamClosed := false
	for {
		select {
		case ev, ok := <-stream:
			if ok {
				log.Printf("Event = %v", ev)
			} else if !streamClosed {
				log.Print("Stream closed")
				streamClosed = true
			}
		case <-cancelTimer.C:
			log.Print("Cancelling")
			cancel()
		case <-timer.C:
			log.Print("Exit")
			return
		}
	}
}

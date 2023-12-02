package main

import (
	"context"
	"log"
	"os"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract"
	"go.uber.org/zap"
)

type LimitsChecker struct {
	ex *exchanges.Exchange
}

func watchChannels(ex exchanges.Exchange) {
	prices, err := ex.WatchSymbolPrice(context.Background(), "BTCUSD")
	if err != nil {
		log.Fatal("prices connect ", err)
	}

	orders, err := ex.WatchOrdersStatuses(context.Background())
	if err != nil {
		log.Fatal("orders connect ", err)
	}

	isFirstOrderShown := false
	for {
		select {
		case priceEv := <-prices:
			switch {
			case priceEv.DisconnectedWithErr != nil:
				log.Fatal("prices disconnect ", priceEv.DisconnectedWithErr)
			case priceEv.Payload != nil:
				log.Print("price ", priceEv.Payload)
			case priceEv.Reconnected != nil:
				log.Print("prices reconnected")
			default:
				panic("")
			}
		case orderEv := <-orders:
			switch {
			case orderEv.DisconnectedWithErr != nil:
				log.Fatal("orders disconnect ", orderEv.DisconnectedWithErr)
			case orderEv.Payload != nil:
				if !isFirstOrderShown {
					log.Print("order ", orderEv.Payload)
				}
				isFirstOrderShown = true
			case orderEv.Reconnected != nil:
				log.Print("orders reconnected")
			default:
				panic("")
			}
		}
	}
}

func main() {
	key, _ := os.LookupEnv("PHEMEX_KEY")
	secret, _ := os.LookupEnv("PHEMEX_SECRET")
	phemex_contract.InitScalesSubscriber(zap.NewExample())

	instances := 12
	lim := phemex_contract.NewPhemexRateLimiter(zap.NewExample())
	for i := 0; i < instances; i++ {
		log.Printf("instance %d", i)

		ex := phemex_contract.NewPhemexContract(key, secret, lim, zap.NewExample())

		ex.StartBackgroundJob(context.Background())

		if i+1 < instances {
			go watchChannels(ex)
		} else {
			watchChannels(ex)
		}
	}
}

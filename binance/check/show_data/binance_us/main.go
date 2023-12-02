package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance"
	"go.uber.org/zap"
)

func main() {
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	ctx := context.Background()
	ex := binance.NewBinanceUS(binance.OriginalBinanceURLs, key, secret, zap.NewExample())

	openOrders, err := ex.GetOpenOrders(ctx)
	printAsJSON(openOrders, err)

	symbol := "BINANCE-LTCUSDT"
	orders, err := ex.GetOrders(ctx, exchanges.OrderFilter{Symbol: &symbol})
	printAsJSON(orders, err)

	account, err := ex.GetAccount(ctx)
	printAsJSON(account, err)

	orderID := "211233700"
	orderFilterByClientOrderId, err := ex.GetOrders(ctx, exchanges.OrderFilter{
		Symbol:  &symbol,
		OrderID: &orderID,
	})
	printAsJSON(orderFilterByClientOrderId, err)

	// Watch Price
	// symbol := "BINANCE-LTCUSDT"
	// orders, err := ex.WatchSymbolPrice(ctx, symbol)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// a := <-orders
	// log.Printf("%v", a)
	// for msg := range orders {
	// 	log.Printf("testtest %v", msg)
	// }
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func printAsJSON(x interface{}, err error) {
	fatalIfErr(err)

	text, err := json.MarshalIndent(x, "", " ")
	if err != nil {
		log.Fatalf("JSON Error: %v", err)
	}
	log.Print(string(text))
}

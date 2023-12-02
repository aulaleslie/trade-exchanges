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
	ex := binance.NewBinanceFutures(binance.TestnetBinanceURLs, key, secret, zap.NewExample())

	openOrders, err := ex.GetOpenOrders(ctx)
	printAsJSON(openOrders, err)

	orders, err := ex.GetOrders(ctx, exchanges.OrderFilter{Symbol: nil})
	printAsJSON(orders, err)

	account, err := ex.GetAccount(ctx)
	printAsJSON(account, err)

	clientOrderID := ""
	symbol := ""
	orderFilterByClientOrderId, err := ex.GetOrders(ctx, exchanges.OrderFilter{
		Symbol:        &symbol,
		ClientOrderID: &clientOrderID,
	})
	printAsJSON(orderFilterByClientOrderId, err)

	generatedID, err := ex.GenerateClientOrderID(ctx, "id")
	printAsJSON(generatedID, err)
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

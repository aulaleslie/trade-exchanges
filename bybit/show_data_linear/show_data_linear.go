package main

import (
	"context"
	"fmt"
	"log"
	"os"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/bybit"
	"github.com/cockroachdb/apd"

	"go.uber.org/zap"
)

type BybitLinear struct {
	bybitLinear *bybit.BybitLinear
}

func main() {
	apiKey, _ := os.LookupEnv("KEY")
	apiSecret, _ := os.LookupEnv("SECRET")
	client := bybit.NewBybitLinear(apiKey, apiSecret, "", zap.NewExample())

	if len(os.Args) < 2 {
		log.Fatal("command required")
	}

	cmd := os.Args[1]
	log.Printf("Trying to use command '%v'", cmd)
	os.Args = os.Args[1:]

	bybit := &BybitLinear{client}
	switch cmd {
	case "getsymbols":
		bybit.GetTradableSymbols()
	case "wsorders":
		bybit.FollowWSOrders()
	case "wsprice":
		bybit.FollowWSPrice()
	case "wsposition":
		bybit.FollowWSPosition()
	case "cancelOrder":
		bybit.CancelOrder()
	case "createorder":
		bybit.CreateOrder()
	case "account":
		bybit.GetAccount()
	case "getOpenOrders":
		bybit.GetOpenOrders()
	case "getOrders":
		bybit.GetOrders()
	case "getPrice":
		bybit.GetPrice()
	case "getOrderInfo":
		bybit.GetOrderInfo()
	case "getOrderInfoByExtID":
		bybit.GetOrderInfoByExternalID()
	default:
		log.Fatalf("Command '%s' not found", cmd)
	}
}

func (m *BybitLinear) GetOrderInfo() {
	account, err := m.bybitLinear.GetOrderInfo(context.Background(), "BTCUSDT", "1433936509473398784", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitLinear) GetOrderInfoByExternalID() {
	account, err := m.bybitLinear.GetOrderInfoByClientOrderID(context.Background(), "BTCUSDT", "16856745525511003", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitLinear) GetOpenOrders() {
	openOrders, err := m.bybitLinear.GetOpenOrders(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(openOrders)
}

func (m *BybitLinear) GetAccount() {
	account, err := m.bybitLinear.GetAccount(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitLinear) GetTradableSymbols() {
	account, err := m.bybitLinear.GetTradableSymbols(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitLinear) CreateOrder() {

	// You can adjust price and qty manually
	price, _, _ := apd.NewFromString("25000")
	qty, _, _ := apd.NewFromString("0.001")
	symbol := "BTCUSDT"
	// You can adjust price and qty manually
	account, err := m.bybitLinear.PlaceBuyOrderV2(context.Background(), false, symbol, price, qty, "", "Limit")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitLinear) GetOrders() {
	symbol := "BTCUSDT"
	filter := exchanges.OrderFilter{
		Symbol: &symbol,
	}
	openOrders, err := m.bybitLinear.GetOrders(context.Background(), filter)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(openOrders)
}

func (m *BybitLinear) GetPrice() {
	symbol := "BYBIT-BTCUSDT"
	price, err := m.bybitLinear.GetPrice(context.Background(), symbol)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(price)
}

func (m *BybitLinear) CancelOrder() {
	symbol := "BTCUSDT"
	orderId := "1415994314883960320"
	err := m.bybitLinear.CancelOrder(context.Background(), symbol, orderId)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(orderId)
}

func (m *BybitLinear) FollowWSOrders() {
	ctx := context.Background()
	log.Printf("Watch orders")

	ch, err := m.bybitLinear.WatchOrdersStatuses(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (m *BybitLinear) FollowWSPosition() {
	ctx := context.Background()
	log.Printf("Watch position")

	ch, err := m.bybitLinear.WatchAccountPositions(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (m *BybitLinear) FollowWSPrice() {
	ctx := context.Background()
	log.Printf("Watch price")

	ch, err := m.bybitLinear.WatchSymbolPrice(ctx, "BTCUSDT")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	a := <-ch
	log.Printf("%v", a)
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

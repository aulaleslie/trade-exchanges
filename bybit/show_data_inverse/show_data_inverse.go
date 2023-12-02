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

type BybitInverse struct {
	bybitInverse *bybit.BybitInverse
}

func main() {
	apiKey, _ := os.LookupEnv("KEY")
	apiSecret, _ := os.LookupEnv("SECRET")
	client := bybit.NewBybitInverse(apiKey, apiSecret, "", zap.NewExample())

	if len(os.Args) < 2 {
		log.Fatal("command required")
	}

	cmd := os.Args[1]
	log.Printf("Trying to use command '%v'", cmd)
	os.Args = os.Args[1:]

	bybit := &BybitInverse{client}
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

func (m *BybitInverse) GetOrderInfo() {
	account, err := m.bybitInverse.GetOrderInfo(context.Background(), "BTCUSDT", "1433936509473398784", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitInverse) GetOrderInfoByExternalID() {
	account, err := m.bybitInverse.GetOrderInfoByClientOrderID(context.Background(), "BTCUSDT", "16856745525511003", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}
func (m *BybitInverse) GetOpenOrders() {
	openOrders, err := m.bybitInverse.GetOpenOrders(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(openOrders)
}

func (m *BybitInverse) GetAccount() {
	account, err := m.bybitInverse.GetAccount(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitInverse) GetTradableSymbols() {
	account, err := m.bybitInverse.GetTradableSymbols(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitInverse) CreateOrder() {

	// You can adjust price and qty manually
	price, _, _ := apd.NewFromString("1.002")
	qty, _, _ := apd.NewFromString("5.8")
	symbol := "BTCUSDT"
	// You can adjust price and qty manually
	account, err := m.bybitInverse.PlaceBuyOrderV2(context.Background(), false, symbol, price, qty, "", "LIMIT")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *BybitInverse) GetOrders() {
	symbol := "BTCUSDT"
	filter := exchanges.OrderFilter{
		Symbol: &symbol,
	}
	openOrders, err := m.bybitInverse.GetOrders(context.Background(), filter)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(openOrders)
}

func (m *BybitInverse) GetPrice() {
	symbol := "BYBIT-BTCUSD"
	price, err := m.bybitInverse.GetPrice(context.Background(), symbol)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(price)
}

func (m BybitInverse) CancelOrder() {
	symbol := "BTCUSDT"
	orderId := "1415994314883960320"
	err := m.bybitInverse.CancelOrder(context.Background(), symbol, orderId)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(orderId)
}

func (m *BybitInverse) FollowWSOrders() {
	ctx := context.Background()
	log.Printf("Watch orders")

	ch, err := m.bybitInverse.WatchOrdersStatuses(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (m *BybitInverse) FollowWSPosition() {
	ctx := context.Background()
	log.Printf("Watch position")

	ch, err := m.bybitInverse.WatchAccountPositions(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (m *BybitInverse) FollowWSPrice() {
	ctx := context.Background()
	log.Printf("Watch price")

	ch, err := m.bybitInverse.WatchSymbolPrice(ctx, "BTCUSDT")
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

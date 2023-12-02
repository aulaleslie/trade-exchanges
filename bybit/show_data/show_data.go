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

type Bybit struct {
	bybitContract *bybit.BybitContract
}

func main() {
	apiKey := "PEDFZUPCDXZVSMLYSQ"
	apiSecret := "NCTQMTIDFFMUXYYQKOVEDSDSEDIUVJLHQPAG"
	client := bybit.NewBybitContract(apiKey, apiSecret, "", zap.NewExample())

	if len(os.Args) < 2 {
		log.Fatal("command required")
	}

	cmd := os.Args[1]
	log.Printf("Trying to use command '%v'", cmd)
	os.Args = os.Args[1:]

	bybit := &Bybit{client}
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

func (m *Bybit) GetOrderInfo() {
	account, err := m.bybitContract.GetOrderInfo(context.Background(), "MATICUSDT", "1501665711982958336", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *Bybit) GetOrderInfoByExternalID() {
	account, err := m.bybitContract.GetOrderInfoByClientOrderID(context.Background(), "BTCUSDT", "1501665711982958336", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}
func (m *Bybit) GetOpenOrders() {
	openOrders, err := m.bybitContract.GetOpenOrders(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for _, openOrder := range openOrders {
		log.Printf("clientOrderID: " + *openOrder.ClientOrderID)
		log.Printf("ID: " + openOrder.ID)
	}
}

func (m *Bybit) GetAccount() {
	account, err := m.bybitContract.GetAccount(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *Bybit) GetTradableSymbols() {
	account, err := m.bybitContract.GetTradableSymbols(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *Bybit) CreateOrder() {

	// You can adjust price and qty manually
	price, _, _ := apd.NewFromString("0.1")
	qty, _, _ := apd.NewFromString("3")
	symbol := "MATICUSDT"
	// You can adjust price and qty manually
	account, err := m.bybitContract.PlaceBuyOrderV2(context.Background(), false, symbol, price, qty, "", "LIMIT")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(account)
}

func (m *Bybit) GetOrders() {
	symbol := "BTCUSDT"
	filter := exchanges.OrderFilter{
		Symbol: &symbol,
	}
	openOrders, err := m.bybitContract.GetOrders(context.Background(), filter)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(openOrders)
}

func (m *Bybit) GetPrice() {
	symbol := "BYBIT-BTCUSDT"
	price, err := m.bybitContract.GetPrice(context.Background(), symbol)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(price)
}

func (m *Bybit) CancelOrder() {
	symbol := "MATICUSDT"
	orderId := "1510344337270249216"
	err := m.bybitContract.CancelOrder(context.Background(), symbol, orderId)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(orderId)
}

func (m *Bybit) FollowWSOrders() {
	ctx := context.Background()
	log.Printf("Watch orders")

	ch, err := m.bybitContract.WatchOrdersStatuses(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (m *Bybit) FollowWSPosition() {
	ctx := context.Background()
	log.Printf("Watch position")

	ch, err := m.bybitContract.WatchAccountPositions(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (m *Bybit) FollowWSPrice() {
	ctx := context.Background()
	log.Printf("Watch price")

	ch, err := m.bybitContract.WatchSymbolPrice(ctx, "BTCUSDT")
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

package main

import (
	"log"
	"time"

	"github.com/aulaleslie/trade-exchanges/tradovate"
)

type Tradovate struct {
	tradovateClient      *tradovate.Client
	tradovateAsyncClient *tradovate.AsyncClient
}

// func main() {
// 	client := tradovate.NewTradovateContract(apiKey, apiSecret, "", zap.NewExample())

// 	if len(os.Args) < 2 {
// 		log.Fatal("command required")
// 	}

// 	cmd := os.Args[1]
// 	log.Printf("Trying to use command '%v'", cmd)
// 	os.Args = os.Args[1:]

// 	tradovate := &Tradovate{client}
// 	switch cmd {
// 	case "getsymbols":
// 		tradovate.GetTradableSymbols()
// 	case "wsorders":
// 		tradovate.FollowWSOrders()
// 	case "wsprice":
// 		tradovate.FollowWSPrice()
// 	case "wsposition":
// 		tradovate.FollowWSPosition()
// 	case "cancelOrder":
// 		tradovate.CancelOrder()
// 	case "createorder":
// 		tradovate.CreateOrder()
// 	case "account":
// 		tradovate.GetAccount()
// 	case "getOpenOrders":
// 		tradovate.GetOpenOrders()
// 	case "getOrders":
// 		tradovate.GetOrders()
// 	case "getPrice":
// 		tradovate.GetPrice()
// 	case "getOrderInfo":
// 		tradovate.GetOrderInfo()
// 	case "getOrderInfoByExtID":
// 		tradovate.GetOrderInfoByExternalID()
// 	default:
// 		log.Fatalf("Command '%s' not found", cmd)
// 	}
// }

func main() {
	client := tradovate.NewClient(
		"demo",
		"Grid Bot Service",
		"1.0",
		"kulsoomkhanani",
		"Desidesi54!",
		"2107",
		"65d2f10d-1d9d-4a4b-872c-8a0a79d80417",
	)

	asyncClient := tradovate.NewAsyncClient(
		"demo",
		"Grid Bot Service",
		"1.0",
		"kulsoomkhanani",
		"Desidesi54!",
		"2107",
		"65d2f10d-1d9d-4a4b-872c-8a0a79d80417",
	)

	// if len(os.Args) < 2 {
	// 	log.Fatal("command required")
	// }

	// cmd := os.Args[1]
	// log.Printf("Trying to use command '%v'", cmd)
	// os.Args = os.Args[1:]

	tradovate := &Tradovate{client, asyncClient}
	tradovate.FollowWSPrice()
	// switch cmd {
	// case "getsymbols":
	// 	tradovate.GetTradableSymbols()
	// case "wsorders":
	// 	tradovate.FollowWSOrders()
	// case "wsprice":
	// 	tradovate.FollowWSPrice()
	// case "wsposition":
	// 	tradovate.FollowWSPosition()
	// case "cancelOrder":
	// 	tradovate.CancelOrder()
	// case "createorder":
	// 	tradovate.CreateOrder()
	// case "account":
	// 	tradovate.GetAccount()
	// case "getOpenOrders":
	// 	tradovate.GetOpenOrders()
	// case "getOrders":
	// 	tradovate.GetOrders()
	// case "getPrice":
	// 	tradovate.GetPrice()
	// case "getOrderInfo":
	// 	tradovate.GetOrderInfo()
	// case "getOrderInfoByExtID":
	// 	tradovate.GetOrderInfoByExternalID()
	// default:
	// 	log.Fatalf("Command '%s' not found", cmd)
	// }
}

func (m *Tradovate) GetOrderInfo() {
	// account, err := m.tradovateContract.GetOrderInfo(context.Background(), "MATICUSDT", "1501665711982958336", nil)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(account)
}

func (m *Tradovate) GetOrderInfoByExternalID() {
	// account, err := m.tradovateContract.GetOrderInfoByClientOrderID(context.Background(), "BTCUSDT", "1501665711982958336", nil)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(account)
}
func (m *Tradovate) GetOpenOrders() {
	// openOrders, err := m.tradovateContract.GetOpenOrders(context.Background())
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(openOrders)
}

func (m *Tradovate) GetAccount() {
	// account, err := m.tradovateContract.GetAccount(context.Background())
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(account)
}

func (m *Tradovate) GetTradableSymbols() {
	// account, err := m.tradovateContract.GetTradableSymbols(context.Background())
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(account)
}

func (m *Tradovate) CreateOrder() {

	// // You can adjust price and qty manually
	// price, _, _ := apd.NewFromString("0.1")
	// qty, _, _ := apd.NewFromString("3")
	// symbol := "MATICUSDT"
	// // You can adjust price and qty manually
	// account, err := m.tradovateContract.PlaceBuyOrderV2(context.Background(), false, symbol, price, qty, "", "LIMIT")
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(account)
}

func (m *Tradovate) GetOrders() {
	// symbol := "BTCUSDT"
	// filter := exchanges.OrderFilter{
	// 	Symbol: &symbol,
	// }
	// openOrders, err := m.tradovateContract.GetOrders(context.Background(), filter)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(openOrders)
}

func (m *Tradovate) GetPrice() {
	// symbol := "TRADOVATE-BTCUSDT"
	// price, err := m.tradovateContract.GetPrice(context.Background(), symbol)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(price)
}

func (m *Tradovate) CancelOrder() {
	// symbol := "MATICUSDT"
	// orderId := "1498761072727873280"
	// err := m.tradovateContract.CancelOrder(context.Background(), symbol, orderId)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// fmt.Println(orderId)
}

func (m *Tradovate) FollowWSOrders() {
	// ctx := context.Background()
	// log.Printf("Watch orders")

	// ch, err := m.tradovateContract.WatchOrdersStatuses(ctx)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// for msg := range ch {
	// 	log.Printf("%v", msg)
	// }
	// log.Printf("loop finished")
}

func (m *Tradovate) FollowWSPosition() {
	// ctx := context.Background()
	// log.Printf("Watch position")

	// ch, err := m.tradovateContract.WatchAccountPositions(ctx)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }
	// for msg := range ch {
	// 	log.Printf("%v", msg)
	// }
	// log.Printf("loop finished")
}

func (m *Tradovate) FollowWSPrice() {

	err := m.tradovateClient.ConnectWebsocket()
	if err != nil {
		log.Fatalf("Error Connect: %v", err)
	}

	go m.tickListener()

	for _, m := range m.tradovateClient.RequestPool {
		for msg := range m {
			log.Printf("%v", msg)
		}
	}

}

func (m *Tradovate) tickListener() {
	log.Println("[tickListener] sleep 20 seconds....")
	time.Sleep(20 * time.Second)
	log.Println("[tickListener] started...")

	m.tradovateClient.GetHistoricalTickData("MNQZ3", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
}

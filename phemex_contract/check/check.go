package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aulaleslie/trade-exchanges/phemex_contract"
	"github.com/aulaleslie/trade-exchanges/utils"
	"go.uber.org/zap"
)

func main() {
	// checkScalesSubscriber()
	// checkPriceSubscriber()
	// checkWatchPrice()
	// checkGetPrice()
	// checkSubscribeOrders()
	// checkListActiveOrders()
	// checkListInactiveOrders()
	// checkCancelOrder()
	// checkGetOrder()
	// checkPlaceOrder()
	//checkFetchAfterPlace()
	checkPlaceOrderV2()
	//checkSubscribePositions()
}

func checkScalesSubscriber() {
	ss := phemex_contract.ScalesSubscriber{}
	go func() {
		time.Sleep(time.Second * 20)
		ss.Start(context.Background())
	}()
	for range time.Tick(time.Second * 5) {
		log.Print(ss.GetLastScales())
	}
}

func checkPriceSubscriber() {
	ch, err := phemex_contract.SubscribeToPrice(context.Background(), "BTCUSD", zap.NewExample())
	if err != nil {
		log.Fatal(err)
	}

	for x := range ch {
		log.Printf("price: %v", x)
	}
}

func checkWatchPrice() {
	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract("", "", phemex_contract.NewPhemexRateLimiter(lg), lg)

	ch, err := phx.WatchSymbolPrice(context.Background(), "BTCUSD")
	if err != nil {
		log.Fatal(err)
	}

	for x := range ch {
		log.Printf("price: %v", x)
	}
}

func checkGetPrice() {
	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract("", "", phemex_contract.NewPhemexRateLimiter(lg), lg)

	time.Sleep(20 * time.Second)

	ctx := context.Background()

	price, err := phx.GetPrice(ctx, "BTCUSD")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("price: %v", price)
}

func checkSubscribeOrders() {
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	ch, err := phx.WatchOrdersStatuses(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// important := map[string]struct{}{}

	for msg := range ch {
		if msg.DisconnectedWithErr != nil {
			log.Fatalf("disconnected: %v", msg.DisconnectedWithErr)
		}

		time.Sleep(1 * time.Second)

		// if _, ok := important[msg.Payload.OrderID]; !ok {
		// 	continue
		// }
		log.Printf("msg: %v, smbl: %s", msg, *msg.Payload.Symbol)
		ctx := context.Background()

		oInfo, err := phx.GetOrderInfo(ctx, *msg.Payload.Symbol, msg.Payload.OrderID, nil)
		if err != nil {
			log.Printf("ERROR: .GetOrderInfo: %v", err)
			log.Printf("SKIP: getOrderInfo by clientOrderID b/c error of getOrder by OrderID")
			continue
		}
		clOID := ""
		if oInfo.ClientOrderID != nil {
			clOID = *oInfo.ClientOrderID
		}
		log.Printf("oInfo: %v, clOID: %s", oInfo, clOID)

		if oInfo.ClientOrderID == nil {
			log.Printf("SKIP: order have no client orderID")
			continue
		}

		cloInfo, err := phx.GetOrderInfoByClientOrderID(ctx,
			*msg.Payload.Symbol, clOID, utils.TimePtr(time.Now().Add(-5*24*time.Hour)))
		if err != nil {
			log.Printf("ERROR: .GetOrderInfoByClientOrderID(%s, %s): %v", *msg.Payload.Symbol, clOID, err)
			continue
		}

		log.Printf("cloInfo: %v, clclOID: %s", cloInfo, *cloInfo.ClientOrderID)
	}
}

func checkSubscribePositions() {
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	ch, err := phx.WatchAccountPositions(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// important := map[string]struct{}{}

	for msg := range ch {
		if msg.DisconnectedWithErr != nil {
			log.Fatalf("disconnected: %v", msg.DisconnectedWithErr)
		}

		time.Sleep(1 * time.Second)
		var positions = msg.Payload
		log.Printf("Position :%f", positions[0].Value)
	}
}

func checkListActiveOrders() {
	// log.Printf("checkListActiveOrders")
	// key, _ := os.LookupEnv("KEY")
	// secret, _ := os.LookupEnv("SECRET")

	// phx := phemex_contract.NewPhemexContract(key, secret)
	// err := phx.GetActiveOrders("BTCUSD")
	// log.Printf("err: %v", err)
}

func checkListInactiveOrders() {
	// log.Printf("checkListInactiveOrders")
	// key, _ := os.LookupEnv("KEY")
	// secret, _ := os.LookupEnv("SECRET")

	// phx := phemex_contract.NewPhemexContract(key, secret)
	// err := phx.GetInactiveOrders("BTCUSD")
	// log.Printf("err: %v", err)
}

func checkCancelOrder() {
	log.Printf("checkCancelOrder")
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	// outdatedSoNotFound := "7c1bc768-0a8c-4e88-bc49-1083cf8284b1"
	// err := phx.CancelOrder("BTCUSD", outdatedSoNotFound)
	// log.Printf("error: %v", err)

	alreadyFilled := "e5717d87-7984-4e78-aacd-c07ebf1706ec"
	// // alreadyFilled := "1a51e04b-9043-4b3a-8a57-988d3a34c6b2"
	ctx := context.Background()
	err := phx.CancelOrder(ctx, "BTCUSD", alreadyFilled)
	log.Printf("error: %v", err)

	// unknown := "1a51e04b-9043-4b3a-8a57-988d3a34c6b8"
	// err := phx.CancelOrder("BTCUSD", unknown)
	// log.Printf("error: %v", err)
}

func checkGetOrder() {
	log.Print("checkGetOrder")
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	// alreadyCanceled := "7c1bc768-0a8c-4e88-bc49-1083cf8284b1"
	// info, err := phx.GetOrderInfo("BTCUSD", alreadyCanceled)
	// log.Printf("info = %v, error: %v", info, err)

	ctx := context.Background()
	filled := "e5717d87-7984-4e78-aacd-c07ebf1706ec"
	info, err := phx.GetOrderInfo(ctx, "BTCUSD", filled, nil)
	log.Printf("info = %v, error: %v", info, err)
}

func checkPlaceOrder() {
	log.Print("checkPlaceOrder")
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	phx.StartBackgroundJob(context.Background())

	symbol := "BTCUSD"
	price := utils.FromString("34000")
	qty := utils.FromString("1")

	err := phemex_contract.ScalesSubscriberInstance.CheckOrUpdate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	priceRound, err := phx.RoundPrice(ctx, symbol, price, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("priceR: %v", priceRound)

	qtyRound, err := phx.RoundQuantity(ctx, symbol, qty)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("qtyR: %v", qtyRound)

	id, err := phx.PlaceBuyOrder(ctx, true, symbol, priceRound, qtyRound, "TEST-first-m")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("id = %s; err=%v", id, err)
}

func checkPlaceOrderV2() {
	log.Print("checkPlaceOrderV2")
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	phx.StartBackgroundJob(context.Background())

	symbol := "BTCUSD"
	price := utils.FromString("34000")
	qty := utils.FromString("1")

	err := phemex_contract.ScalesSubscriberInstance.CheckOrUpdate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	priceRound, err := phx.RoundPrice(ctx, symbol, price, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("priceR: %v", priceRound)

	qtyRound, err := phx.RoundQuantity(ctx, symbol, qty)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("qtyR: %v", qtyRound)

	id, err := phx.PlaceBuyOrderV2(ctx, true, symbol, priceRound, qtyRound, "TEST-first-m", "Stop")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("id = %s; err=%v", id, err)
}

func checkFetchAfterPlace() {
	log.Print("checkFetchAfterPlace")
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	phx := phemex_contract.NewPhemexContract(key, secret, phemex_contract.NewPhemexRateLimiter(lg), lg)

	phx.StartBackgroundJob(context.Background())

	symbol := "PHX-BTCUSD"
	price := utils.FromString("30500")
	qty := utils.FromString("1")

	err := phemex_contract.ScalesSubscriberInstance.CheckOrUpdate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	priceRound, err := phx.RoundPrice(ctx, symbol, price, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("priceR: %v", priceRound)

	qtyRound, err := phx.RoundQuantity(ctx, symbol, qty)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("qtyR: %v", qtyRound)

	id, err := phx.PlaceBuyOrder(ctx, true, symbol, priceRound, qtyRound, "TEST-second-h")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Printf("Placed: id = %s", id)
	}()

	// for {
	// 	order, err := phx.GetOrderInfo(symbol, id)
	// 	if err == nil {
	// 		log.Printf("Fetched ok: %v", order)
	// 		break
	// 	}
	// 	log.Printf("Fetch error: %v", err)
	// 	time.Sleep(100 * time.Millisecond)
	// }

	for {
		err := phx.CancelOrder(ctx, symbol, id)
		if err == nil {
			log.Printf("Cancelled successfully")
			break
		}
		log.Printf("Cancel error: %v", err)
		time.Sleep(100 * time.Millisecond)
	}
}

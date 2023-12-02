package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	api "github.com/adshao/go-binance/v2"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance"
	"github.com/cockroachdb/apd"
	"go.uber.org/zap"
)

type Binance struct {
	client *api.Client
	ex     exchanges.Exchange
}

func main() {
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	client := api.NewClient(key, secret)

	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())

	if len(os.Args) < 2 {
		log.Fatal("command required")
	}

	cmd := os.Args[1]
	log.Printf("Trying to use command '%v'", cmd)
	os.Args = os.Args[1:]

	bn := &Binance{client, ex}
	switch cmd {
	case "account":
		bn.PrintAccount()
	case "orders":
		bn.PrintOrders()
	case "wsorders":
		bn.FollowWSOrders()
	case "wsprice":
		bn.FollowWSPrice()
	case "cancel":
		bn.CancelOrder()
	case "createorder":
		bn.CancelOrder()
	default:
		log.Fatalf("Command '%s' not found", cmd)
	}
}

func (b *Binance) CreateOrder() {
	symbolPtr := flag.String("symbol", "", "symbol")
	flag.Parse()

	ctx := context.Background()

	// You can adjust price and qty manually
	price, _, _ := apd.NewFromString("1800")
	qty, _, _ := apd.NewFromString("0.02")
	// You can adjust price and qty manually

	id, err := b.ex.PlaceBuyOrderV2(ctx, true, *symbolPtr, price, qty, "", "LIMIT")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Order Created Order ID: %v", id)
}

func (b *Binance) PrintAccount() {
	ctx := context.Background()
	account, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	balances := account.Balances
	account.Balances = nil
	printAsJSON(account, nil)

	log.Print("Balances: [")
	replacer := strings.NewReplacer("0", "", ".", "")
	for _, coin := range balances {
		if replacer.Replace(coin.Free) == "" &&
			replacer.Replace(coin.Locked) == "" {
			continue
		}
		log.Printf(" %s: {free: %s, locked: %s}", coin.Asset, coin.Free, coin.Locked)
	}
	log.Print("]")
}

func (b *Binance) CancelOrder() {
	symbolPtr := flag.String("symbol", "", "symbol")
	clientOrderIDPtr := flag.String("coid", "", "clientOrderID")
	flag.Parse()

	if *symbolPtr == "" || *clientOrderIDPtr == "" {
		log.Fatalf(
			`Error some is empty: symbol = "%s" || clientOrderID == "%s"`,
			*symbolPtr, *clientOrderIDPtr)
	}

	log.Printf(`symbol = "%s"`, *symbolPtr)
	log.Printf(`clientOrderID == "%s"`, *clientOrderIDPtr)

	ctx := context.Background()
	_, err := b.client.NewCancelOrderService().
		Symbol(*symbolPtr).
		OrigClientOrderID(*clientOrderIDPtr).
		Do(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Print("Canceled")
}

func (b *Binance) PrintOrders() {
	symbolPtr := flag.String("symbol", "", "symbol")
	fromDayPtr := flag.String("from-day", "", "from day in format YY-MM-DD")
	flag.Parse()
	symbol := strings.ToUpper(*symbolPtr)
	fromTime := time.Time{}
	if *fromDayPtr != "" {
		// Mon Jan 2 15:04:05 MST 2006
		var err error
		fromTime, err = time.Parse("06-01-02", *fromDayPtr)
		fatalIfErr(err)
	}

	log.Print(os.Args)
	log.Printf("Symbol: %v", symbol)
	log.Printf("From time: %v", fromTime)

	b.printOrdersInternal(symbol, fromTime)
}

func (b *Binance) printOrdersInternal(symbol string, fromTime time.Time) {
	ctx := context.Background()
	svc := b.client.NewListOrdersService()
	svc.Symbol(symbol)
	if (fromTime != time.Time{}) {
		svc.StartTime(fromTime.UnixNano() / 1e6)
	}

	orders, err := svc.Do(ctx)
	fatalIfErr(err)

	bytes, err := json.Marshal(orders)
	fatalIfErr(err)

	data := []map[string]interface{}{}
	err = json.Unmarshal(bytes, &data)
	fatalIfErr(err)

	for _, order := range data {
		t := order["time"].(float64)
		order["time"] = time.Unix(0, int64(t)*1e6).String()

		ut := order["updateTime"].(float64)
		order["updateTime"] = time.Unix(0, int64(ut)*1e6).String()
	}

	printAsJSON(data, err)
}

func (b *Binance) FollowWSOrders() {
	symbolPtr := flag.String("symbol", "", "symbol")
	flag.Parse()
	symbol := strings.ToUpper(*symbolPtr)
	log.Printf("Symbol: %v", symbol)

	ctx := context.Background()
	events, err := b.ex.WatchOrdersStatuses(ctx)
	fatalIfErr(err)

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	for ev := range events {
		log.Print(strings.Repeat("=", 60))
		log.Printf("> ev: %v", ev)
		fatalIfErr(ev.DisconnectedWithErr)

		b.printOrdersInternal(symbol, startOfDay)
	}
}

func (b *Binance) FollowWSPrice() {
	symbolPtr := flag.String("symbol", "", "symbol")
	flag.Parse()
	symbol := strings.ToUpper(*symbolPtr)
	log.Printf("Symbol: %v", symbol)

	if symbol == "" {
		log.Fatal("Can't work without symbol")
	}

	ctx := context.Background()
	events, err := b.ex.WatchSymbolPrice(ctx, symbol)
	fatalIfErr(err)

	for ev := range events {
		log.Print(strings.Repeat("=", 60))
		log.Printf("> ev: %v", ev)
		fatalIfErr(ev.DisconnectedWithErr)
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

func fatalIfErr(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

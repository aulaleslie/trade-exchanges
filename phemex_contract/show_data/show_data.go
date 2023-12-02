package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	api "github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract"
	forkAPI "github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"go.uber.org/zap"
)

type ShowData struct {
	client               *api.Client
	forkClient           *forkAPI.Client
	combinedOrderFetcher *phemex_contract.CombinedOrdersFetcher
	positionFetcher      *phemex_contract.PositionFetcher
	ex                   exchanges.Exchange
}

func main() {
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	lg := zap.NewExample()
	client := api.NewClient(key, secret)
	forkClient := forkAPI.NewClient(key, secret, lg)
	rateLim := phemex_contract.NewPhemexRateLimiter(lg)
	combinedOrdersFetcher := phemex_contract.NewCombinedOrdersFetcher(
		client, forkClient,
		rateLim,
		lg,
	)
	positionFetcher := phemex_contract.NewPositionFetcher(client)

	ex := phemex_contract.NewPhemexContract(key, secret, rateLim, zap.NewExample())

	sd := &ShowData{client, forkClient, combinedOrdersFetcher, positionFetcher, ex}
	cmds := map[string]func(){
		"account":         sd.PrintAccount,
		"get-order":       sd.GetOrder,
		"get-order-cof":   sd.GetOrderCOF,
		"get-order-info":  sd.GetOrderInfo,
		"open-orders":     sd.PrintOpenOrders,
		"closed-orders":   sd.PrintClosedOrders,
		"trades":          sd.PrintTrades,
		"symbols":         sd.PrintTradableSymbols,
		"round-qty":       sd.RoundQuantity,
		"round-price":     sd.RoundPrice,
		"watch-orders":    sd.WatchOrders,
		"watch-price":     sd.WatchPrice,
		"cancel":          sd.CancelOrder,
		"get-orders":      sd.GetOrders,
		"get-open-orders": sd.GetOpenOrders,
		"get-account":     sd.GetAccount,
	}

	if len(os.Args) < 2 {
		log.Printf("command required")
		log.Printf("Available commands:")
		for c := range cmds {
			fmt.Printf("%s, ", c)
		}
		os.Exit(1)
	}

	cmd := os.Args[1]
	if fn, ok := cmds[cmd]; ok {
		log.Printf("Trying to use command '%v'", cmd)
		os.Args = os.Args[1:]
		// sd.startBackground()
		fn()
		return
	}

	log.Printf("Command '%s' not found", cmd)
	log.Printf("Available commands:")
	for c := range cmds {
		fmt.Printf("%s, ", c)
	}
	os.Exit(1)
}

func (sd *ShowData) GetOrders() {
	// NOT WORKING YET
	ctx := context.Background()
	res, err := sd.combinedOrderFetcher.GetHistoryOrders(ctx, "BTCUSD", nil, nil)
	printAsJSON(res, err)

	orderID := "1b4f8545-da9e-4cda-b1d1-7a7c91afdd4d"
	resFilterByOrderID, err := sd.combinedOrderFetcher.GetHistoryOrders(ctx, "BTCUSD", &orderID, nil)
	printAsJSON(resFilterByOrderID, err)
}

func (sd *ShowData) GetAccount() {
	ctx := context.Background()
	res, err := sd.positionFetcher.GetAccountPosition(ctx)
	printAsJSON(res, err)
}

func (sd *ShowData) GetOpenOrders() {
	ctx := context.Background()
	res, err := sd.combinedOrderFetcher.GetOpenOrders(ctx)
	printAsJSON(res, err)
}

func (sd *ShowData) startBackground() {
	phemex_contract.InitScalesSubscriber(zap.NewExample())
	err := phemex_contract.ScalesSubscriberInstance.CheckOrUpdate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if sd.client.APIKey == "" {
		return
	}

	defer time.Sleep(time.Second * 5)

	err = sd.combinedOrderFetcher.Start(context.Background())
	if err != nil {
		log.Printf("cof.Start err: %v", err)
	}

	err = sd.ex.(*phemex_contract.PhemexContract).StartBackgroundJob(context.Background())
	if err != nil {
		log.Printf("ex.StartBackground err: %v", err)
	}
}

func (sd *ShowData) PrintAccount() {
	ctx := context.Background()

	printAccount := func(currency string) {
		apData, err := sd.client.NewGetAccountPositionService().Currency(currency).Do(ctx)
		fatalIfErr(err)

		log.Printf("%s Account:", currency)
		printAsJSON(apData.Account, nil)
	}
	printAccount("BTC")
	printAccount("USD")
}

func (sd *ShowData) PrintOpenOrders() {
	ctx := context.Background()

	symbolPtr := flag.String("symbol", "", "symbol")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatal("Can't work without symbol")
	}
	log.Printf("Using %s symbol", *symbolPtr)

	openOrders, rateLimHeaders, err := sd.forkClient.NewListOpenOrdersService().Symbol(*symbolPtr).Do(ctx)
	printAsJSON(rateLimHeaders, nil)
	fatalIfErr(err)

	printAsJSON(openOrders, nil)
}

type orderWithTime struct {
	ActionTime   time.Time
	TransactTime time.Time
	Price        *apd.Decimal
	Order        *forkAPI.OrderResponse
}

func (sd *ShowData) PrintClosedOrders() {
	ctx := context.Background()

	symbolPtr := flag.String("symbol", "", "symbol")
	fromTimePtr := flag.String("from-time", "", "from day in format 'YY-MM-DD HH:mm'")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatal("Can't work without symbol")
	}
	fromTime := time.Time{}
	if *fromTimePtr != "" {
		// Mon Jan 2 15:04:05 MST 2006
		var err error
		fromTime, err = time.ParseInLocation("06-01-02 15:04", *fromTimePtr, time.Now().Location())
		fatalIfErr(err)
	}

	log.Printf("Symbol: %v", *symbolPtr)
	log.Printf("From time: %v", fromTime)

	symbol := phemex_contract.ToPhemexSymbol(*symbolPtr)

	err := phemex_contract.ScalesSubscriberInstance.CheckOrUpdate(ctx)
	fatalIfErr(err)

	symbolScales, err := phemex_contract.ScalesSubscriberInstance.GetLastSymbolScales(symbol)
	fatalIfErr(err)

	limit := 200
	fullList := []*orderWithTime{}
	for offset := 0; true; offset += limit {
		query := sd.forkClient.NewListInactiveOrdersService().
			Symbol(symbol).
			Limit(limit).
			Offset(offset)

		if fromTime != (time.Time{}) {
			query = query.Start(fromTime)
		}

		list, rateLimHeaders, err := query.Do(ctx)
		printAsJSON(rateLimHeaders, nil)
		fatalIfErr(err)

		for _, order := range list {
			fullList = append(fullList, &orderWithTime{
				Order:        order,
				ActionTime:   time.Unix(0, order.ActionTimeNs),
				TransactTime: time.Unix(0, order.TransactTimeNs),
				Price:        utils.Div(apd.New(order.PriceEp, 0), symbolScales.PriceScaleDivider),
			})
		}

		if len(list) < limit {
			break
		}
	}

	sort.Sort(orderWithTimeSort(fullList))
	for _, order := range fullList {
		printAsJSON(order, nil)
	}
}

func (sd *ShowData) PrintTrades() {
	ctx := context.Background()

	symbolPtr := flag.String("symbol", "", "symbol")
	fromTimePtr := flag.String("from-time", "", "from day in format 'YY-MM-DD HH:mm'")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatal("Can't work without symbol")
	}
	fromTime := time.Time{}
	if *fromTimePtr != "" {
		// Mon Jan 2 15:04:05 MST 2006
		var err error
		fromTime, err = time.ParseInLocation("06-01-02 15:04", *fromTimePtr, time.Now().Location())
		fatalIfErr(err)
	}

	log.Printf("Symbol: %v", *symbolPtr)
	log.Printf("From time: %v", fromTime)

	symbol := phemex_contract.ToPhemexSymbol(*symbolPtr)

	limit := 200
	{
		query := sd.forkClient.NewTradesService().
			Symbol(symbol).
			Limit(limit).
			Offset(0)

		if fromTime != (time.Time{}) {
			query = query.Start(fromTime)
		}

		resp, rateLimHeaders, err := query.Do(ctx)
		printAsJSON(rateLimHeaders, nil)
		printAsJSON(resp, err)
	}
}

func (sd *ShowData) PrintTradableSymbols() {
	ctx := context.Background()

	symbols, err := sd.ex.GetTradableSymbols(ctx)
	printAsJSON(symbols, err)
}

func (sd *ShowData) WatchOrders() {
	ctx := context.Background()

	log.Printf("Watch orders")

	ch, err := sd.ex.WatchOrdersStatuses(ctx)
	fatalIfErr(err)

	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (sd *ShowData) WatchPrice() {
	ctx := context.Background()

	symbolPtr := flag.String("symbol", "", "symbol")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatal("Can't work without symbol")
	}
	log.Printf("Using %s symbol", *symbolPtr)

	symbol := phemex_contract.ToPhemexSymbol(*symbolPtr)
	ch, err := sd.ex.WatchSymbolPrice(ctx, symbol)
	fatalIfErr(err)

	for msg := range ch {
		log.Printf("%v", msg)
	}
	log.Printf("loop finished")
}

func (sd *ShowData) RoundPrice() {
	ctx := context.Background()

	symbolPtr := flag.String("symbol", "", "symbol")
	pricePtr := flag.String("price", "", "price")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatal("Can't work without symbol")
	}
	if *pricePtr == "" {
		log.Fatal("Can't work without price")
	}
	price := utils.FromString(*pricePtr)
	log.Printf("Using %s symbol", *symbolPtr)
	log.Printf("Using %v price", price)

	rounded, err := sd.ex.RoundPrice(ctx, *symbolPtr, price, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("RoundedPrice = %v", rounded)
}

func (sd *ShowData) RoundQuantity() {
	ctx := context.Background()

	symbolPtr := flag.String("symbol", "", "symbol")
	qtyPtr := flag.String("qty", "", "qty")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatal("Can't work without 'symbol'")
	}
	if *qtyPtr == "" {
		log.Fatal("Can't work without 'qty'")
	}
	qty := utils.FromString(*qtyPtr)
	log.Printf("Using %s symbol", *symbolPtr)
	log.Printf("Using %v quantity", qty)

	rounded, err := sd.ex.RoundQuantity(ctx, *symbolPtr, qty)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("RoundedQty = %v", rounded)
}

func (sd *ShowData) CancelOrder() {
	symbolPtr := flag.String("symbol", "", "symbol")
	orderIDPtr := flag.String("id", "", "orderID")
	flag.Parse()

	if *symbolPtr == "" || *orderIDPtr == "" {
		log.Fatalf(
			`Error some is empty: symbol = "%s" || orderID == "%s"`,
			*symbolPtr, *orderIDPtr)
	}

	log.Printf(`symbol = "%s"`, *symbolPtr)
	log.Printf(`orderID == "%s"`, *orderIDPtr)

	ctx := context.Background()
	symbol := phemex_contract.ToPhemexSymbol(*symbolPtr)

	err := sd.ex.CancelOrder(ctx, symbol, *orderIDPtr)
	fatalIfErr(err)
	log.Print("Canceled")
}

func (sd *ShowData) GetOrder() {
	symbolPtr := flag.String("symbol", "", "symbol")
	orderIDPtr := flag.String("id", "", "orderID")
	clientOrderIDPtr := flag.String("clid", "", "clientOrderID")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatalf(`Error symbol is empty`)
	}

	if *orderIDPtr == "" && *clientOrderIDPtr == "" {
		log.Fatalf(
			`Error some is empty: orderID == "%s" || clientOrderID == "%s"`,
			*orderIDPtr, *clientOrderIDPtr)
	}

	log.Printf(`symbol = "%s"`, *symbolPtr)
	log.Printf(`orderID == "%s"`, *orderIDPtr)
	log.Printf(`clientOrderID == "%s"`, *clientOrderIDPtr)

	symbol := phemex_contract.ToPhemexSymbol(*symbolPtr)
	query := sd.forkClient.NewQueryOrderService().Symbol(symbol)
	if *orderIDPtr != "" {
		query = query.OrderID(*orderIDPtr)
	}
	if *clientOrderIDPtr != "" {
		query = query.ClOrderID(*clientOrderIDPtr)
	}

	result, headers, err := query.Do(context.Background())
	printAsJSON(headers, nil)
	printAsJSON(result, err)
}

func (sd *ShowData) GetOrderCOF() {
	symbolPtr := flag.String("symbol", "", "symbol")
	orderIDPtr := flag.String("id", "", "orderID")
	clientOrderIDPtr := flag.String("clid", "", "clientOrderID")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatalf(`Error symbol is empty`)
	}

	if *orderIDPtr == "" && *clientOrderIDPtr == "" {
		log.Fatalf(
			`Error some is empty: orderID == "%s" || clientOrderID == "%s"`,
			*orderIDPtr, *clientOrderIDPtr)
	}

	log.Printf(`symbol = "%s"`, *symbolPtr)
	log.Printf(`orderID == "%s"`, *orderIDPtr)
	log.Printf(`clientOrderID == "%s"`, *clientOrderIDPtr)

	ctx := context.Background()
	symbol := phemex_contract.ToPhemexSymbol(*symbolPtr)

	if *orderIDPtr != "" {
		oi, or, err := sd.combinedOrderFetcher.GetOrderInfoByOrderID(ctx, symbol, *orderIDPtr)
		fatalIfErr(err)
		log.Print("orderInfo:")
		printAsString(oi, nil)
		log.Print("orderResponse:")
		printAsString(or, nil)
	}
	if *clientOrderIDPtr != "" {
		oi, or, err := sd.combinedOrderFetcher.GetOrderInfoByClientOrderID(ctx, symbol, *clientOrderIDPtr)
		fatalIfErr(err)
		log.Print("orderInfo:")
		printAsString(oi, nil)
		log.Print("orderResponse:")
		printAsString(or, nil)
	}
}

func (sd *ShowData) GetOrderInfo() {
	log.Printf("GetOrderInfo")
	symbolPtr := flag.String("symbol", "", "symbol")
	orderIDPtr := flag.String("id", "", "orderID")
	clientOrderIDPtr := flag.String("clid", "", "clientOrderID")
	flag.Parse()

	if *symbolPtr == "" {
		log.Fatalf(`Error symbol is empty`)
	}

	if *orderIDPtr == "" && *clientOrderIDPtr == "" {
		log.Fatalf(
			`Error some is empty: orderID == "%s" || clientOrderID == "%s"`,
			*orderIDPtr, *clientOrderIDPtr)
	}

	log.Printf(`symbol = "%s"`, *symbolPtr)
	log.Printf(`orderID == "%s"`, *orderIDPtr)
	log.Printf(`clientOrderID == "%s"`, *clientOrderIDPtr)

	ctx := context.Background()

	if *orderIDPtr != "" {
		t := time.Now()
		printAsString(sd.ex.GetOrderInfo(ctx, *symbolPtr, *orderIDPtr, &t))
	}
	if *clientOrderIDPtr != "" {
		t := time.Now()
		printAsString(sd.ex.GetOrderInfoByClientOrderID(ctx, *symbolPtr, *clientOrderIDPtr, &t))
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

func printAsString(x interface{}, err error) {
	fatalIfErr(err)

	log.Printf("%v", x)
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

////////////////

type orderWithTimeSort []*orderWithTime

func (x orderWithTimeSort) Len() int {
	return len(x)
}

func (x orderWithTimeSort) Less(i, j int) bool {
	if (x[i].ActionTime).Equal(x[j].ActionTime) {
		return x[i].Order.OrderID < x[j].Order.OrderID
	}
	return (x[i].ActionTime).Before(x[j].ActionTime)
}

func (x orderWithTimeSort) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

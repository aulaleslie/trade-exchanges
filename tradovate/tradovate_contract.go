package tradovate

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"net/url"
// 	"regexp"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/cockroachdb/apd"
// 	"github.com/fatih/structs"
// 	tradovate "github.com/hirokisan/bybit/v2"
// 	"github.com/pkg/errors"
// 	exchanges "github.com/aulaleslie/trade-exchanges"
// 	"github.com/aulaleslie/trade-exchanges/utils"
// 	"go.uber.org/zap"
// )

// const TRADOVATE_PREFIX = "TRADOVATE-"

// type TradovateContract struct {
// 	client     *tradovate.Client
// 	wsClient   *tradovate.WebSocketClient
// 	httpClient *http.Client
// 	lg         *zap.Logger
// 	key        string
// 	secret     string
// }

// // TODO: don't forget to check time
// var _ exchanges.Exchange = (*TradovateContract)(nil) // Type check

// func NewTradovateContract(apiKey, secretKey, host string, lg *zap.Logger) *TradovateContract {
// 	lg = lg.Named("Tradovate")

// 	t := &TradovateContract{}

// 	t.client = NewTradovateRestClient(apiKey, secretKey, lg)
// 	t.wsClient = NewTradovateWSClient(apiKey, secretKey, lg)
// 	t.httpClient = NewHTTPClient(time.Second * 10)
// 	t.key = apiKey
// 	t.secret = secretKey
// 	t.lg = lg

// 	return t
// }

// func ToTradovateFullSymbol(symbol string) string {
// 	return TRADOVATE_PREFIX + symbol
// }

// func ToTradovateSymbol(symbol string) string {
// 	return strings.TrimPrefix(symbol, TRADOVATE_PREFIX)
// }

// func (t *TradovateContract) GetOrders(ctx context.Context, filter exchanges.OrderFilter) (res []exchanges.OrderDetailInfo, err error) {
// 	var symbol *tradovate.SymbolV5

// 	if filter.Symbol != nil {
// 		tradovateSymbol := ToTradovateSymbol(*filter.Symbol)
// 		symbol = (*tradovate.SymbolV5)(&tradovateSymbol)
// 	}

// 	params := tradovate.V5GetHistoryOrdersParam{
// 		Symbol:   symbol,
// 		Category: tradovate.CategoryV5Spot,
// 		OrderID:  filter.OrderID,
// 	}

// 	historyOrdersResponse, err := t.client.V5().Order().GetHistoryOrders(params)

// 	if err != nil {
// 		return
// 	}

// 	for _, historyOrder := range historyOrdersResponse.Result.List {
// 		price := utils.FromString(historyOrder.Price)
// 		quantity := utils.FromString(historyOrder.Qty)
// 		executedQuantity := utils.Sub(quantity, utils.FromString(historyOrder.LeavesQty))

// 		fmt.Println(price, executedQuantity)

// 		orderStatusType := mapOrderStatusType(string(historyOrder.OrderStatus))
// 		orderSide := mapOrderSide(string(historyOrder.Side))
// 		orderType := mapOrderType(string(historyOrder.OrderType))

// 		createdTime, err := strconv.ParseInt(historyOrder.CreatedTime, 10, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		orderDetailInfo := exchanges.OrderDetailInfo{
// 			ID:            historyOrder.OrderID,
// 			Symbol:        string(historyOrder.Symbol),
// 			ClientOrderID: &historyOrder.OrderLinkID,
// 			Price:         price,
// 			Quantity:      quantity,
// 			ExecutedQty:   executedQuantity,
// 			Status:        orderStatusType,
// 			OrderType:     orderType,
// 			Time:          createdTime,
// 			OrderSide:     orderSide,
// 			TimeInForce:   nil,
// 			StopPrice:     nil,
// 			QuoteQuantity: nil,
// 		}

// 		res = append(res, orderDetailInfo)
// 	}

// 	return
// }

// func (t *TradovateContract) GetOpenOrders(ctx context.Context) (res []exchanges.OrderDetailInfo, err error) {
// 	input := tradovate.V5GetOpenOrdersParam{
// 		Category: tradovate.CategoryV5Spot,
// 	}

// 	openOrders, err := t.client.V5().Order().GetOpenOrders(input)
// 	if err != nil {
// 		return
// 	}

// 	// TODO: marshal map
// 	for _, openOrder := range openOrders.Result.List {
// 		price := utils.FromString(openOrder.Price)
// 		quantity := utils.FromString(openOrder.Qty)
// 		executedQuantity := utils.Sub(quantity, utils.FromString(openOrder.LeavesQty))

// 		fmt.Println(price, executedQuantity)

// 		orderStatusType := mapOrderStatusType(string(openOrder.OrderStatus))
// 		orderSide := mapOrderSide(string(openOrder.Side))
// 		orderType := mapOrderType(string(openOrder.OrderType))

// 		createdTime, err := strconv.ParseInt(openOrder.CreatedTime, 10, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		orderDetailInfo := exchanges.OrderDetailInfo{
// 			ID:            openOrder.OrderID,
// 			Symbol:        string(openOrder.Symbol),
// 			ClientOrderID: &openOrder.OrderLinkID,
// 			Price:         price,
// 			Quantity:      quantity,
// 			ExecutedQty:   executedQuantity,
// 			Status:        orderStatusType,
// 			OrderType:     orderType,
// 			Time:          createdTime,
// 			OrderSide:     orderSide,
// 			TimeInForce:   nil,
// 			StopPrice:     nil,
// 			QuoteQuantity: nil,
// 		}

// 		res = append(res, orderDetailInfo)
// 	}

// 	return
// }

// func (t *TradovateContract) GetAccount(ctx context.Context) (res exchanges.Account, err error) {
// 	balances, err := t.client.V5().Account().GetWalletBalance(tradovate.AccountType(tradovate.AccountTypeV5UNIFIED), nil)
// 	if err != nil {
// 		return res, errors.Wrap(err, "unable to do GetAccount request")
// 	}

// 	accountBalances := make([]exchanges.AccountBalance, 0)
// 	for _, balance := range balances.Result.List[0].Coin {
// 		free := utils.FromString(balance.WalletBalance)
// 		locked := utils.FromString(balance.Locked)
// 		accountBalances = append(accountBalances, exchanges.AccountBalance{
// 			Coin:   string(balance.Coin),
// 			Free:   free,
// 			Locked: locked,
// 		})
// 	}
// 	res.AccountBalances = accountBalances
// 	return res, nil
// }

// func (t *TradovateContract) RoundPrice(_ context.Context, symbol string, price *apd.Decimal, tickSize *string) (*apd.Decimal, error) {
// 	// TODO: handle this more accurately at bot side
// 	str := price.Text('f')
// 	rgx := priceFloorRE
// 	if tickSize != nil {
// 		precision := utils.FindPrecisionFromTickSize(*tickSize)
// 		if precision != nil {
// 			rgx = regexp.MustCompile(fmt.Sprintf(`^[0-9]{1,20}(\.[0-9]{1,%v})?`, *precision))
// 		}
// 	}
// 	substr := rgx.FindString(str)
// 	if substr == "" {
// 		return nil, errors.Errorf("invalid price %v", price)
// 	}
// 	return utils.FromStringErr(substr)
// }

// func (t *TradovateContract) RoundQuantity(_ context.Context, symbol string, qty *apd.Decimal) (*apd.Decimal, error) {
// 	// TODO: handle this more accurately at bot side
// 	str := qty.Text('f')
// 	substr := qtyFloorRE.FindString(str)
// 	if substr == "" {
// 		return nil, errors.Errorf("invalid quantity %v", qty)
// 	}
// 	return utils.FromStringErr(str)
// }

// func (t *TradovateContract) GetPrefix() string {
// 	return TRADOVATE_PREFIX
// }

// func (t *TradovateContract) GetName() string {
// 	return "Tradovate"
// }

// func (t *TradovateContract) PlaceBuyOrder(ctx context.Context,
// 	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
// ) (id string, e error) {
// 	return t.PlaceBuyOrderV2(ctx, true, symbol, price, quantity, prefferedID, string(tradovate.OrderTypeLimit))
// }

// func (t *TradovateContract) PlaceSellOrder(ctx context.Context,
// 	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
// ) (id string, e error) {
// 	return t.PlaceSellOrderV2(ctx, true, symbol, price, quantity, prefferedID, string(tradovate.OrderTypeLimit))
// }

// func (t *TradovateContract) CancelOrder(ctx context.Context, symbol, id string) error {
// 	input := tradovate.V5CancelOrderParam{
// 		Symbol:   tradovate.SymbolV5(symbol),
// 		OrderID:  &id,
// 		Category: tradovate.CategoryV5Spot,
// 	}
// 	_, err := t.client.V5().Order().CancelOrder(input)
// 	if err != nil {
// 		return errors.Wrap(err, "unable to do CancelOrder on orderId ")
// 	}

// 	return nil
// }

// func (t *TradovateContract) ReleaseOrder(_ context.Context, symbol, id string) error {
// 	// _, err := b.client.V5().Order().
// 	return nil
// }

// func (t *TradovateContract) GetOrderInfo(ctx context.Context, symbol, id string, _ *time.Time) (exchanges.OrderInfo, error) {

// 	orderInfo := exchanges.OrderInfo{}

// 	baseURL := TradovateBaseURL + GetOrderHistoryPath
// 	params := url.Values{}
// 	params.Set("category", string(tradovate.CategoryV5Spot))
// 	params.Set("symbol", symbol)
// 	params.Set("orderId", id)

// 	reqURL, err := url.Parse(baseURL)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to create http request")
// 	}
// 	reqURL.RawQuery = params.Encode()

// 	signature, timestamp := tradovateSignatureGenerator(t.key, t.secret, reqURL.RawQuery)

// 	req, err := http.NewRequest("GET", reqURL.String(), nil)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to create http request")
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
// 	req.Header.Set("X-BAPI-SIGN", signature)
// 	req.Header.Set("X-BAPI-API-KEY", t.key)
// 	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
// 	req.Header.Set("X-BAPI-RECV-WINDOW", recWindow)

// 	resp, err := t.httpClient.Do(req)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to get order history")
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to read response order history")
// 	}

// 	var orderRespon TradovateResponse
// 	err = json.Unmarshal(body, &orderRespon)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to unmarshall response order history")
// 	}
// 	fmt.Println(orderRespon)

// 	if len(orderRespon.Result.List) == 0 {
// 		return orderInfo, errors.Wrap(err, "order not found")
// 	}
// 	orderInfo.ID = orderRespon.Result.List[0].OrderId
// 	orderInfo.ClientOrderID = &orderRespon.Result.List[0].OrderLinkId
// 	orderInfo.Status = mapOrderStatusType(string(orderRespon.Result.List[0].OrderStatus))

// 	return orderInfo, nil

// }

// func (t *TradovateContract) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string, _ *time.Time) (exchanges.OrderInfo, error) {
// 	orderInfo := exchanges.OrderInfo{}

// 	baseURL := TradovateBaseURL + GetOrderHistoryPath
// 	params := url.Values{}
// 	params.Set("category", string(tradovate.CategoryV5Spot))
// 	params.Set("symbol", symbol)
// 	params.Set("orderLinkId", clientOrderID)

// 	reqURL, err := url.Parse(baseURL)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to create http request")
// 	}
// 	reqURL.RawQuery = params.Encode()

// 	signature, timestamp := tradovateSignatureGenerator(t.key, t.secret, reqURL.RawQuery)

// 	req, err := http.NewRequest("GET", reqURL.String(), nil)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to create http request")
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
// 	req.Header.Set("X-BAPI-SIGN", signature)
// 	req.Header.Set("X-BAPI-API-KEY", t.key)
// 	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
// 	req.Header.Set("X-BAPI-RECV-WINDOW", recWindow)

// 	resp, err := t.httpClient.Do(req)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to get order history")
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to read response order history")
// 	}

// 	var orderRespon TradovateResponse
// 	err = json.Unmarshal(body, &orderRespon)
// 	if err != nil {
// 		return orderInfo, errors.Wrap(err, "unable to unmarshall response order history")
// 	}
// 	fmt.Println(orderRespon)

// 	if len(orderRespon.Result.List) == 0 {
// 		return orderInfo, errors.Wrap(err, "order not found")
// 	}
// 	orderInfo.ID = orderRespon.Result.List[0].OrderId
// 	orderInfo.ClientOrderID = &orderRespon.Result.List[0].OrderLinkId
// 	orderInfo.Status = mapOrderStatusType(string(orderRespon.Result.List[0].OrderStatus))

// 	return orderInfo, nil
// }

// func (t *TradovateContract) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
// 	tradovateSmbl := ToTradovateSymbol(symbol)

// 	symbl := tradovate.SymbolV5(tradovateSmbl)
// 	input := tradovate.V5GetTickersParam{
// 		Symbol:   &symbl,
// 		Category: tradovate.CategoryV5Spot,
// 	}
// 	res, err := t.client.V5().Market().GetTickers(input)
// 	if err != nil {
// 		return utils.Zero, err
// 	}
// 	price := utils.FromString(res.Result.Spot.List[0].LastPrice)
// 	return price, nil
// }

// type Server struct {
// 	Name        string
// 	ID          int
// 	Enabled     bool
// 	http.Server // embedded
// }

// func (t *TradovateContract) GetTradableSymbols(ctx context.Context) ([]exchanges.SymbolInfo, error) {
// 	input := tradovate.V5GetInstrumentsInfoParam{
// 		Category: tradovate.CategoryV5Spot,
// 	}
// 	res, err := t.client.V5().Market().GetInstrumentsInfo(input)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "unable to do GetTradableSymbols request")
// 	}

// 	result := []exchanges.SymbolInfo{}
// 	for _, symbol := range res.Result.Spot.List {
// 		fullSymbol := ToTradovateFullSymbol(string(symbol.Symbol))
// 		f := &SymbolFilter{
// 			LotSizeFilter: symbol.LotSizeFilter,
// 			PriceFilter:   symbol.PriceFilter,
// 		}
// 		fltrMap := structs.Map(f)
// 		sInfo := exchanges.SymbolInfo{
// 			DisplayName:    fullSymbol,
// 			Symbol:         fullSymbol,
// 			OriginalSymbol: string(symbol.Symbol),
// 			Filters:        []map[string]interface{}{fltrMap},
// 		}
// 		result = append(result, sInfo)
// 	}
// 	return result, nil
// }

// // WatchOrdersStatuses Returns control immediately
// func (t *TradovateContract) WatchOrdersStatuses(ctx context.Context) (<-chan exchanges.OrderEvent, error) {
// 	return SubscribeToOrders(ctx, t.wsClient, t.lg, tradovate.CategoryV5Spot)
// }

// // // WatchOrdersStatuses Returns control immediately
// // func (b *TradovateContract) WatchFills(ctx context.Context) (<-chan exchanges.FillsEvent, error) {
// // 	// 	wsEndpoint := rest.WS_ENDPOINT
// // 	// 	if strings.Contains(b.client.Host, rest.ENDPOINT_US) {
// // 	// 		wsEndpoint = rest.WS_ENDPOINT_US
// // 	// 	}

// // 	// 	return SubscribeToFills(ctx, wsEndpoint, b.key, b.secret, b.lg)
// // 	return nil, nil

// // }

// // WatchSymbolPrice
// func (t *TradovateContract) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan exchanges.PriceEvent, error) {
// 	tradovateSymbol := ToTradovateSymbol(symbol)
// 	in, err := SubscribeToPrices(ctx, t.wsClient, t.lg, tradovateSymbol)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return in, nil
// }

// // PlaceBuyOrderV2 Place Buy Order with OrderType param
// func (t *TradovateContract) PlaceBuyOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
// 	qttString := utils.ToFlatString(qty)
// 	priceString := utils.ToFlatString(price)
// 	symbol = ToTradovateSymbol(symbol)
// 	input := tradovate.V5CreateOrderParam{
// 		Category:    tradovate.CategoryV5Spot,
// 		Symbol:      tradovate.SymbolV5(symbol),
// 		Side:        tradovate.SideBuy,
// 		OrderType:   tradovate.OrderType(orderType),
// 		Qty:         qttString,
// 		Price:       &priceString,
// 		OrderLinkID: &preferredID,
// 	}
// 	res, err := t.client.V5().Order().CreateOrder(input)
// 	if err != nil {
// 		return "", errors.Wrap(err, "unable to do PlaceBuyOrderV2 request")
// 	}

// 	return res.Result.OrderID, nil
// }

// // PlaceSellOrderV2 Place Sell Order with OrderType param
// func (t *TradovateContract) PlaceSellOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
// 	qttString := utils.ToFlatString(qty)
// 	priceString := utils.ToFlatString(price)
// 	symbol = ToTradovateSymbol(symbol)
// 	input := tradovate.V5CreateOrderParam{
// 		Category:    tradovate.CategoryV5Spot,
// 		Symbol:      tradovate.SymbolV5(symbol),
// 		Side:        tradovate.SideSell,
// 		OrderType:   tradovate.OrderType(orderType),
// 		Qty:         qttString,
// 		Price:       &priceString,
// 		OrderLinkID: &preferredID,
// 	}
// 	res, err := t.client.V5().Order().CreateOrder(input)
// 	if err != nil {
// 		return "", errors.Wrap(err, "unable to do PlaceSellOrderV2 request")
// 	}
// 	return res.Result.OrderID, nil

// }

// func (t *TradovateContract) WatchAccountPositions(ctx context.Context) (<-chan exchanges.PositionEvent, error) {
// 	in, err := SubscribeToPositions(ctx, t.wsClient, t.lg, tradovate.CategoryV5Spot)
// 	if err != nil {
// 		return nil, err
// 	}

// 	out := make(chan exchanges.PositionEvent, 100)
// 	go func() {
// 		defer close(out)
// 		for ev := range in {
// 			out <- ev
// 		}
// 	}()
// 	return out, nil
// }

// func (t *TradovateContract) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
// 	return utils.GenClientOrderID(identifierID)
// }

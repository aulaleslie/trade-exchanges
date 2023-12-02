package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/fatih/structs"
	"github.com/hirokisan/bybit/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const BYBIT_PREFIX = "BYBIT-"

type BybitContract struct {
	client     *bybit.Client
	wsClient   *bybit.WebSocketClient
	httpClient *http.Client
	lg         *zap.Logger
	key        string
	secret     string
}

// TODO: don't forget to check time
var _ exchanges.Exchange = (*BybitContract)(nil) // Type check

func NewBybitContract(apiKey, secretKey, host string, lg *zap.Logger) *BybitContract {
	lg = lg.Named("Bybit")

	b := &BybitContract{}

	b.client = NewBybitRestClient(apiKey, secretKey, lg)
	b.wsClient = NewBybitWSClient(apiKey, secretKey, lg)
	b.httpClient = NewHTTPClient(time.Second * 10)
	b.key = apiKey
	b.secret = secretKey
	b.lg = lg

	return b
}

func ToBybitFullSymbol(symbol string) string {
	return BYBIT_PREFIX + symbol
}

func ToBybitSymbol(symbol string) string {
	return strings.TrimPrefix(symbol, BYBIT_PREFIX)
}

func (b *BybitContract) GetOrders(ctx context.Context, filter exchanges.OrderFilter) (res []exchanges.OrderDetailInfo, err error) {
	var symbol *bybit.SymbolV5

	if filter.Symbol != nil {
		bybitSymbol := ToBybitSymbol(*filter.Symbol)
		symbol = (*bybit.SymbolV5)(&bybitSymbol)
	}

	params := bybit.V5GetHistoryOrdersParam{
		Symbol:   symbol,
		Category: bybit.CategoryV5Spot,
		OrderID:  filter.OrderID,
	}

	historyOrdersResponse, err := b.client.V5().Order().GetHistoryOrders(params)

	if err != nil {
		return
	}

	for _, historyOrder := range historyOrdersResponse.Result.List {
		price := utils.FromString(historyOrder.Price)
		quantity := utils.FromString(historyOrder.Qty)
		executedQuantity := utils.Sub(quantity, utils.FromString(historyOrder.LeavesQty))

		fmt.Println(price, executedQuantity)

		orderStatusType := mapOrderStatusType(string(historyOrder.OrderStatus))
		orderSide := mapOrderSide(string(historyOrder.Side))
		orderType := mapOrderType(string(historyOrder.OrderType))

		createdTime, err := strconv.ParseInt(historyOrder.CreatedTime, 10, 64)
		if err != nil {
			return nil, err
		}
		orderDetailInfo := exchanges.OrderDetailInfo{
			ID:            historyOrder.OrderID,
			Symbol:        string(historyOrder.Symbol),
			ClientOrderID: &historyOrder.OrderLinkID,
			Price:         price,
			Quantity:      quantity,
			ExecutedQty:   executedQuantity,
			Status:        orderStatusType,
			OrderType:     orderType,
			Time:          createdTime,
			OrderSide:     orderSide,
			TimeInForce:   nil,
			StopPrice:     nil,
			QuoteQuantity: nil,
		}

		res = append(res, orderDetailInfo)
	}

	return
}

func (b *BybitContract) GetOpenOrders(ctx context.Context) (res []exchanges.OrderDetailInfo, err error) {
	input := bybit.V5GetOpenOrdersParam{
		Category: bybit.CategoryV5Spot,
	}

	openOrders, err := b.client.V5().Order().GetOpenOrders(input)
	if err != nil {
		return
	}

	// TODO: marshal map
	for _, openOrder := range openOrders.Result.List {
		price := utils.FromString(openOrder.Price)
		quantity := utils.FromString(openOrder.Qty)
		executedQuantity := utils.Sub(quantity, utils.FromString(openOrder.LeavesQty))

		orderLinkID := openOrder.OrderLinkID

		orderStatusType := mapOrderStatusType(string(openOrder.OrderStatus))
		orderSide := mapOrderSide(string(openOrder.Side))
		orderType := mapOrderType(string(openOrder.OrderType))

		createdTime, err := strconv.ParseInt(openOrder.CreatedTime, 10, 64)
		if err != nil {
			return nil, err
		}
		orderDetailInfo := exchanges.OrderDetailInfo{
			ID:            openOrder.OrderID,
			Symbol:        string(openOrder.Symbol),
			ClientOrderID: &orderLinkID,
			Price:         price,
			Quantity:      quantity,
			ExecutedQty:   executedQuantity,
			Status:        orderStatusType,
			OrderType:     orderType,
			Time:          createdTime,
			OrderSide:     orderSide,
			TimeInForce:   nil,
			StopPrice:     nil,
			QuoteQuantity: nil,
		}

		res = append(res, orderDetailInfo)
	}

	return
}

func (b *BybitContract) GetAccount(ctx context.Context) (res exchanges.Account, err error) {
	balances, err := b.client.V5().Account().GetWalletBalance(bybit.AccountType(bybit.AccountTypeV5UNIFIED), nil)
	if err != nil {
		return res, errors.Wrap(err, "unable to do GetAccount request")
	}

	accountBalances := make([]exchanges.AccountBalance, 0)
	for _, balance := range balances.Result.List[0].Coin {
		free := utils.FromString(balance.WalletBalance)
		locked := utils.FromString(balance.Locked)
		accountBalances = append(accountBalances, exchanges.AccountBalance{
			Coin:   string(balance.Coin),
			Free:   free,
			Locked: locked,
		})
	}
	res.AccountBalances = accountBalances
	return res, nil
}

func (b *BybitContract) RoundPrice(_ context.Context, symbol string, price *apd.Decimal, tickSize *string) (*apd.Decimal, error) {
	// TODO: handle this more accurately at bot side
	str := price.Text('f')
	rgx := priceFloorRE
	if tickSize != nil {
		precision := utils.FindPrecisionFromTickSize(*tickSize)
		if precision != nil {
			rgx = regexp.MustCompile(fmt.Sprintf(`^[0-9]{1,20}(\.[0-9]{1,%v})?`, *precision))
		}
	}
	substr := rgx.FindString(str)
	if substr == "" {
		return nil, errors.Errorf("invalid price %v", price)
	}
	return utils.FromStringErr(substr)
}

func (b *BybitContract) RoundQuantity(_ context.Context, symbol string, qty *apd.Decimal) (*apd.Decimal, error) {
	// TODO: handle this more accurately at bot side
	str := qty.Text('f')
	substr := qtyFloorRE.FindString(str)
	if substr == "" {
		return nil, errors.Errorf("invalid quantity %v", qty)
	}
	return utils.FromStringErr(str)
}

func (b *BybitContract) GetPrefix() string {
	return BYBIT_PREFIX
}

func (b *BybitContract) GetName() string {
	return "Bybit Spot"
}

func (b *BybitContract) PlaceBuyOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	return b.PlaceBuyOrderV2(ctx, true, symbol, price, quantity, prefferedID, string(bybit.OrderTypeLimit))
}

func (b *BybitContract) PlaceSellOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	return b.PlaceSellOrderV2(ctx, true, symbol, price, quantity, prefferedID, string(bybit.OrderTypeLimit))
}

func (b *BybitContract) CancelOrder(ctx context.Context, symbol, id string) error {
	input := bybit.V5CancelOrderParam{
		Symbol:      bybit.SymbolV5(ToBybitSymbol(symbol)),
		OrderLinkID: &id,
		Category:    bybit.CategoryV5Spot,
	}

	_, err := b.client.V5().Order().CancelOrder(input)
	if err != nil {
		err = nil
		// try to cancel with order id
		input = bybit.V5CancelOrderParam{
			Symbol:   bybit.SymbolV5(ToBybitSymbol(symbol)),
			OrderID:  &id,
			Category: bybit.CategoryV5Spot,
		}

		_, err = b.client.V5().Order().CancelOrder(input)
		if err != nil {
			return errors.Wrap(err, "unable to do CancelOrder on orderId ")
		}
	}

	return nil
}

func (b *BybitContract) ReleaseOrder(_ context.Context, symbol, id string) error {
	// _, err := b.client.V5().Order().
	return nil
}

func (b *BybitContract) GetOrderInfo(ctx context.Context, symbol, id string, _ *time.Time) (exchanges.OrderInfo, error) {
	return b.GetOrderInfoByClientOrderID(ctx, symbol, id, nil)
}

func (b *BybitContract) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string, _ *time.Time) (exchanges.OrderInfo, error) {
	orderInfo := exchanges.OrderInfo{}

	baseURL := BybitBaseURL + GetOrderHistoryPath
	params := url.Values{}
	params.Set("category", string(bybit.CategoryV5Spot))
	params.Set("symbol", symbol)
	params.Set("orderLinkId", clientOrderID)

	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return orderInfo, errors.Wrap(err, "unable to create http request")
	}
	reqURL.RawQuery = params.Encode()

	signature, timestamp := bybitSignatureGenerator(b.key, b.secret, reqURL.RawQuery)

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return orderInfo, errors.Wrap(err, "unable to create http request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("X-BAPI-API-KEY", b.key)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-BAPI-RECV-WINDOW", recWindow)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return orderInfo, errors.Wrap(err, "unable to get order history")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return orderInfo, errors.Wrap(err, "unable to read response order history")
	}

	var orderRespon BybitResponse
	err = json.Unmarshal(body, &orderRespon)
	if err != nil {
		return orderInfo, errors.Wrap(err, "unable to unmarshall response order history")
	}
	fmt.Println(orderRespon)

	if len(orderRespon.Result.List) == 0 {
		return orderInfo, errors.Wrap(err, "order not found")
	}
	orderInfo.ID = orderRespon.Result.List[0].OrderId
	orderInfo.ClientOrderID = &orderRespon.Result.List[0].OrderLinkId
	orderInfo.Status = mapOrderStatusType(string(orderRespon.Result.List[0].OrderStatus))

	return orderInfo, nil
}

func (b *BybitContract) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
	bybitSmbl := ToBybitSymbol(symbol)

	symbl := bybit.SymbolV5(bybitSmbl)
	input := bybit.V5GetTickersParam{
		Symbol:   &symbl,
		Category: bybit.CategoryV5Spot,
	}
	res, err := b.client.V5().Market().GetTickers(input)
	if err != nil {
		return utils.Zero, err
	}
	price := utils.FromString(res.Result.Spot.List[0].LastPrice)
	return price, nil
}

type Server struct {
	Name        string
	ID          int
	Enabled     bool
	http.Server // embedded
}

func (b *BybitContract) GetTradableSymbols(ctx context.Context) ([]exchanges.SymbolInfo, error) {
	input := bybit.V5GetInstrumentsInfoParam{
		Category: bybit.CategoryV5Spot,
	}
	res, err := b.client.V5().Market().GetInstrumentsInfo(input)
	if err != nil {
		return nil, errors.Wrap(err, "unable to do GetTradableSymbols request")
	}

	result := []exchanges.SymbolInfo{}
	for _, symbol := range res.Result.Spot.List {
		fullSymbol := ToBybitFullSymbol(string(symbol.Symbol))
		f := &SymbolFilter{
			LotSizeFilter: symbol.LotSizeFilter,
			PriceFilter:   symbol.PriceFilter,
		}
		fltrMap := structs.Map(f)
		sInfo := exchanges.SymbolInfo{
			DisplayName:    fullSymbol,
			Symbol:         fullSymbol,
			OriginalSymbol: string(symbol.Symbol),
			Filters:        []map[string]interface{}{fltrMap},
		}
		result = append(result, sInfo)
	}
	return result, nil
}

// WatchOrdersStatuses Returns control immediately
func (b *BybitContract) WatchOrdersStatuses(ctx context.Context) (<-chan exchanges.OrderEvent, error) {
	return SubscribeToOrders(ctx, b.wsClient, b.lg, bybit.CategoryV5Spot)
}

// // WatchOrdersStatuses Returns control immediately
// func (b *BybitContract) WatchFills(ctx context.Context) (<-chan exchanges.FillsEvent, error) {
// 	// 	wsEndpoint := rest.WS_ENDPOINT
// 	// 	if strings.Contains(b.client.Host, rest.ENDPOINT_US) {
// 	// 		wsEndpoint = rest.WS_ENDPOINT_US
// 	// 	}

// 	// 	return SubscribeToFills(ctx, wsEndpoint, b.key, b.secret, b.lg)
// 	return nil, nil

// }

// WatchSymbolPrice
func (b *BybitContract) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan exchanges.PriceEvent, error) {
	bybitSymbol := ToBybitSymbol(symbol)
	in, err := SubscribeToPrices(ctx, b.wsClient, b.lg, bybitSymbol)
	if err != nil {
		return nil, err
	}

	return in, nil
}

// PlaceBuyOrderV2 Place Buy Order with OrderType param
func (b *BybitContract) PlaceBuyOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	qttString := utils.ToFlatString(qty)
	priceString := utils.ToFlatString(price)
	symbol = ToBybitSymbol(symbol)
	input := bybit.V5CreateOrderParam{
		Category:    bybit.CategoryV5Spot,
		Symbol:      bybit.SymbolV5(symbol),
		Side:        bybit.SideBuy,
		OrderType:   bybit.OrderType(orderType),
		Qty:         qttString,
		Price:       &priceString,
		OrderLinkID: &preferredID,
	}
	res, err := b.client.V5().Order().CreateOrder(input)
	if err != nil {
		return "", errors.Wrap(err, "unable to do PlaceBuyOrderV2 request")
	}

	return res.Result.OrderLinkID, nil
}

// PlaceSellOrderV2 Place Sell Order with OrderType param
func (b *BybitContract) PlaceSellOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	qttString := utils.ToFlatString(qty)
	priceString := utils.ToFlatString(price)
	symbol = ToBybitSymbol(symbol)
	input := bybit.V5CreateOrderParam{
		Category:    bybit.CategoryV5Spot,
		Symbol:      bybit.SymbolV5(symbol),
		Side:        bybit.SideSell,
		OrderType:   bybit.OrderType(orderType),
		Qty:         qttString,
		Price:       &priceString,
		OrderLinkID: &preferredID,
	}
	res, err := b.client.V5().Order().CreateOrder(input)
	if err != nil {
		return "", errors.Wrap(err, "unable to do PlaceSellOrderV2 request")
	}
	return res.Result.OrderLinkID, nil

}

func (b *BybitContract) WatchAccountPositions(ctx context.Context) (<-chan exchanges.PositionEvent, error) {
	in, err := SubscribeToPositions(ctx, b.wsClient, b.lg, bybit.CategoryV5Spot)
	if err != nil {
		return nil, err
	}

	out := make(chan exchanges.PositionEvent, 100)
	go func() {
		defer close(out)
		for ev := range in {
			out <- ev
		}
	}()
	return out, nil
}

func (b *BybitContract) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
	return utils.GenClientOrderID(identifierID)
}

package binance

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	api "github.com/adshao/go-binance/v2"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	BINANCE_SPOT_LINK_ID_PREFIX = "x-INHON5QW"
)

// BinanceLong TODO: probably it's better to switch to orderID instead of clientOrderID
// Order ID in the methods input and output is clientOrderID internally.
type BinanceLong struct {
	client         *api.Client
	canceller      *BinanceOrderCanceller
	orderGetter    *OrderGetter
	orderPlacer    *OrderPlacer
	positionGetter *PositionGetter
	urls           BinanceURLs
	lg             *zap.Logger
}

// TODO: don't forget to check time
var _ exchanges.Exchange = (*BinanceLong)(nil) // Type check

func NewBinanceLong(urls BinanceURLs, apiKey, secretKey string, lg *zap.Logger) *BinanceLong {
	lg = lg.Named("Binance")

	b := &BinanceLong{}

	b.client = NewBinanceClient(urls.APIURL, apiKey, secretKey, lg)
	b.canceller = NewBinanceOrderCanceller(b.client)
	b.orderGetter = &OrderGetter{b.client}
	b.orderPlacer = NewOrderPlacer(b.client)
	b.positionGetter = &PositionGetter{b.client}
	b.urls = urls
	b.lg = lg
	return b
}

func ToBinanceSymbol(symbol string) string {
	return strings.TrimPrefix(symbol, BINANCE_PREFIX)
}

func ToFullSymbol(binanceSymbol string) string {
	return BINANCE_PREFIX + binanceSymbol
}

func (b *BinanceLong) RoundPrice(_ context.Context, symbol string, price *apd.Decimal, tickSize *string) (*apd.Decimal, error) {
	// TODO: handle this more accurately at bot side
	str := price.Text('f')
	rgx := binancePriceFloorRE
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

func (b *BinanceLong) RoundQuantity(_ context.Context, symbol string, qty *apd.Decimal) (*apd.Decimal, error) {
	// TODO: handle this more accurately at bot side
	str := qty.Text('f')
	substr := binanceQtyFloorRE.FindString(str)
	if substr == "" {
		return nil, errors.Errorf("invalid quantity %v", qty)
	}
	return utils.FromStringErr(str)
}

func (b *BinanceLong) GetPrefix() string {
	return BINANCE_PREFIX
}

func (b *BinanceLong) GetName() string {
	return "Binance Spot"
}

func (b *BinanceLong) PlaceBuyOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrder(ctx, binanceSymbol, price, quantity, prefferedID, api.SideTypeBuy)
}

func (b *BinanceLong) PlaceSellOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrder(ctx, binanceSymbol, price, quantity, prefferedID, api.SideTypeSell)
}

func (b *BinanceLong) CancelOrder(ctx context.Context, symbol, id string) error {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.canceller.CancelOrder(ctx, binanceSymbol, id)
}

func (b *BinanceLong) ReleaseOrder(_ context.Context, symbol, id string) error {
	return nil
}

func (b *BinanceLong) GetOrderInfo(ctx context.Context, symbol, id string, _ *time.Time) (exchanges.OrderInfo, error) {
	return b.GetOrderInfoByClientOrderID(ctx, symbol, id, nil)
}

func (b *BinanceLong) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string, _ *time.Time) (exchanges.OrderInfo, error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderGetter.GetOrderInfoByClientOrderID(ctx, binanceSymbol, clientOrderID)
}

func (b *BinanceLong) GetOpenOrders(ctx context.Context) ([]exchanges.OrderDetailInfo, error) {
	return b.orderGetter.GetOpenOrders(ctx)
}

func (b *BinanceLong) GetOrders(ctx context.Context, filter exchanges.OrderFilter) (res []exchanges.OrderDetailInfo, err error) {
	if filter.Symbol == nil {
		return res, errors.New("symbol is empty!")
	}
	binanceSymbol := ToBinanceSymbol(*filter.Symbol)
	return b.orderGetter.GetHistoryOrders(
		ctx,
		binanceSymbol,
		nil,
		filter.ClientOrderID,
	)
}

func (b *BinanceLong) GetAccount(ctx context.Context) (exchanges.Account, error) {
	return b.positionGetter.GetAccountBalances(ctx)
}

func (b *BinanceLong) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
	binanceSymbol := ToBinanceSymbol(symbol)

	result, err := b.client.NewListPriceChangeStatsService().Symbol(binanceSymbol).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't do request")
	}
	if len(result) == 0 {
		return nil, errors.New("empty result")
	}
	stats := result[0]
	if stats.Symbol != binanceSymbol {
		return nil, errors.Errorf("got result for another symbol: %s", ToFullSymbol(stats.Symbol))
	}
	price, _, err := apd.NewFromString(stats.LastPrice)
	if err != nil {
		return nil, errors.Wrapf(err, "can't convert price from '%s'", stats.LastPrice)
	}
	return price, nil
}

func (b *BinanceLong) GetTradableSymbols(ctx context.Context) ([]exchanges.SymbolInfo, error) {
	info, err := b.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to do ExchangeInfo request")
	}

	result := []exchanges.SymbolInfo{}
	for _, symbol := range info.Symbols {
		if symbol.IsSpotTradingAllowed {
			fullSymbol := ToFullSymbol(symbol.Symbol)
			sInfo := exchanges.SymbolInfo{
				DisplayName:    fullSymbol,
				Symbol:         fullSymbol,
				OriginalSymbol: symbol.Symbol,
				Filters:        symbol.Filters,
			}
			result = append(result, sInfo)
		}
	}
	return result, nil
}

// WatchOrdersStatuses Returns control immediately
func (b *BinanceLong) WatchOrdersStatuses(ctx context.Context) (<-chan exchanges.OrderEvent, error) {
	return SubscribeToOrders(ctx, b.urls, b.client, b.lg)
}

// WatchSymbolPrice OPTIMIZATION: subscribe to single symbol on client side not to all symbols.
func (b *BinanceLong) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan exchanges.PriceEvent, error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return SubscribeToPrice(ctx, b.urls, binanceSymbol, b.lg)
}

// PlaceBuyOrderV2 Place Buy Order with OrderType param
func (b *BinanceLong) PlaceBuyOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrderV2(ctx, binanceSymbol, price, qty, preferredID, api.SideTypeBuy, api.OrderType(orderType))
}

// PlaceSellOrderV2 Place Sell Order with OrderType param
func (b *BinanceLong) PlaceSellOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrderV2(ctx, binanceSymbol, price, qty, preferredID, api.SideTypeSell, api.OrderType(orderType))
}

func (b *BinanceLong) WatchAccountPositions(ctx context.Context) (<-chan exchanges.PositionEvent, error) {
	listenKey, err := b.client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't take listen key")
	}

	wsEndpoint := b.urls.WSUserDataURL(listenKey)

	return SubscribeToPositions(ctx, wsEndpoint, b.lg)
}

func (b *BinanceLong) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
	generatedID, err := utils.GenClientOrderID(identifierID)
	if err != nil {
		return "", err
	}
	return BINANCE_SPOT_LINK_ID_PREFIX + "_" + generatedID, nil
}

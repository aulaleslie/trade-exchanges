package binance

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	api "github.com/adshao/go-binance/v2/futures"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance/futures"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	BINANCE_FUTURES_LINK_ID_PREFIX = "x-9Oc4JrZL"
)

// BinanceLong TODO: probably it's better to switch to orderID instead of clientOrderID
// Order ID in the methods input and output is clientOrderID internally.
type BinanceFutures struct {
	Client         *api.Client
	canceller      *futures.BinanceOrderCanceller
	orderGetter    *futures.OrderGetter
	orderPlacer    *futures.OrderPlacer
	positionGetter *futures.PositionGetter
	urls           BinanceURLs
	lg             *zap.Logger
}

// TODO: don't forget to check time
var _ exchanges.Exchange = (*BinanceFutures)(nil) // Type check

func NewBinanceFutures(urls BinanceURLs, apiKey, secretKey string, lg *zap.Logger) *BinanceFutures {
	lg = lg.Named("BinanceFutures")

	b := &BinanceFutures{}

	b.Client = NewBinanceFuturesClient(urls.FutureAPIURL, apiKey, secretKey, lg)
	b.canceller = futures.NewBinanceOrderCanceller(b.Client)
	b.orderGetter = futures.NewOrderGetter(b.Client)
	b.orderPlacer = futures.NewOrderPlacer(b.Client)
	b.positionGetter = futures.NewPositionGetter(b.Client)
	b.urls = urls
	b.lg = lg
	return b
}

func (b *BinanceFutures) RoundPrice(
	_ context.Context,
	symbol string,
	price *apd.Decimal,
	tickSize *string,
) (*apd.Decimal, error) {
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

func (b *BinanceFutures) RoundQuantity(_ context.Context, symbol string, qty *apd.Decimal) (*apd.Decimal, error) {
	// TODO: handle this more accurately at bot side
	str := qty.Text('f')
	substr := binanceQtyFloorRE.FindString(str)
	if substr == "" {
		return nil, errors.Errorf("invalid quantity %v", qty)
	}
	return utils.FromStringErr(str)
}

func (b *BinanceFutures) GetPrefix() string {
	return BINANCE_PREFIX
}

func (b *BinanceFutures) GetName() string {
	return "Binance Futures"
}

func (b *BinanceFutures) PlaceBuyOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrder(ctx, binanceSymbol, price, quantity, prefferedID, api.SideTypeBuy)
}

func (b *BinanceFutures) PlaceSellOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrder(ctx, binanceSymbol, price, quantity, prefferedID, api.SideTypeSell)
}

func (b *BinanceFutures) CancelOrder(ctx context.Context, symbol, id string) error {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.canceller.CancelOrder(ctx, binanceSymbol, id)
}

func (b *BinanceFutures) ReleaseOrder(_ context.Context, symbol, id string) error {
	return nil
}

func (b *BinanceFutures) GetOrderInfo(ctx context.Context, symbol, id string, _ *time.Time) (exchanges.OrderInfo, error) {
	return b.GetOrderInfoByClientOrderID(ctx, symbol, id, nil)
}

func (b *BinanceFutures) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string, _ *time.Time) (exchanges.OrderInfo, error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderGetter.GetOrderInfoByClientOrderID(ctx, binanceSymbol, clientOrderID)
}

func (b *BinanceFutures) GetOpenOrders(ctx context.Context) ([]exchanges.OrderDetailInfo, error) {
	return b.orderGetter.GetOpenOrders(ctx)
}

func (b *BinanceFutures) GetOrders(ctx context.Context, filter exchanges.OrderFilter) ([]exchanges.OrderDetailInfo, error) {
	return b.orderGetter.GetHistoryOrders(
		ctx,
		filter.Symbol,
		nil,
		filter.ClientOrderID,
	)
}

func (b *BinanceFutures) GetAccount(ctx context.Context) (exchanges.Account, error) {
	return b.positionGetter.GetAccountPosition(ctx)
}

func (b *BinanceFutures) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
	binanceSymbol := ToBinanceSymbol(symbol)

	result, err := b.Client.NewListPriceChangeStatsService().Symbol(binanceSymbol).Do(ctx)
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

func (b *BinanceFutures) GetTradableSymbols(ctx context.Context) ([]exchanges.SymbolInfo, error) {
	info, err := b.Client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to do ExchangeInfo request")
	}

	result := []exchanges.SymbolInfo{}
	for _, symbol := range info.Symbols {
		fullSymbol := ToFullSymbol(symbol.Symbol)
		sInfo := exchanges.SymbolInfo{
			DisplayName:    fullSymbol,
			Symbol:         fullSymbol,
			OriginalSymbol: symbol.Symbol,
			Filters:        symbol.Filters,
		}
		result = append(result, sInfo)
	}
	return result, nil
}

func (b *BinanceFutures) WatchOrdersStatuses(ctx context.Context) (<-chan exchanges.OrderEvent, error) {
	listenKey, err := b.Client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't take listen key")
	}

	// Add capability to keep alive the listen key
	ticker := time.NewTicker(20 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			err := b.Client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
			if err != nil {
				b.lg.Sugar().Errorf("Error to keep alive user stream websocket.. wait until next ticker... reason..", err)
			}
		}
	}()

	wsEndpoint := b.urls.WSFuturesUserDataURL(listenKey)

	if strings.Contains(wsEndpoint, b.urls.FutureWebSocketBaseURL) {
		return SubscribeToOrdersFutures(ctx, wsEndpoint, b.lg)
	}

	return SubscribeToOrdersV2(ctx, wsEndpoint, b.lg)
}

func (b *BinanceFutures) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan exchanges.PriceEvent, error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	b.lg.Sugar().Infof("ws address : %v", b.urls.WSFuturesAllMiniMarketStatsURL())
	return SubscribeToPriceV2(ctx, b.urls.WSFuturesAllMiniMarketStatsURL(), binanceSymbol, b.lg)
}

// PlaceBuyOrderV2 Place Buy Order with OrderType param
func (b *BinanceFutures) PlaceBuyOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrderV2(ctx, binanceSymbol, price, qty, preferredID, api.SideTypeBuy, api.OrderType(orderType))
}

// PlaceSellOrderV2 Place Sell Order with OrderType param
func (b *BinanceFutures) PlaceSellOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrderV2(ctx, binanceSymbol, price, qty, preferredID, api.SideTypeSell, api.OrderType(orderType))
}

func (b *BinanceFutures) WatchAccountPositions(ctx context.Context) (<-chan exchanges.PositionEvent, error) {
	listenKey, err := b.Client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't take listen key")
	}

	wsEndpoint := b.urls.WSFuturesUserDataURL(listenKey)

	if strings.Contains(wsEndpoint, b.urls.FutureWebSocketBaseURL) {
		return SubscribeToPositionsFutures(ctx, wsEndpoint, b.lg)
	}

	return SubscribeToPositions(ctx, wsEndpoint, b.lg)
}

func (b *BinanceFutures) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
	generatedID, err := utils.GenClientOrderID(identifierID)
	if err != nil {
		return "", err
	}
	return BINANCE_FUTURES_LINK_ID_PREFIX + "_" + generatedID, nil
}

package binance

import (
	"context"
	"fmt"
	"regexp"
	"time"

	api "github.com/adshao/go-binance/v2"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// BinanceLong TODO: probably it's better to switch to orderID instead of clientOrderID
// Order ID in the methods input and output is clientOrderID internally.
type BinanceUS struct {
	Client         *api.Client
	canceller      *BinanceOrderCanceller
	orderGetter    *OrderGetter
	orderPlacer    *OrderPlacer
	positionGetter *PositionGetter
	urls           BinanceURLs
	lg             *zap.Logger
}

// TODO: don't forget to check time
var _ exchanges.Exchange = (*BinanceUS)(nil) // Type check

func NewBinanceUS(urls BinanceURLs, apiKey, secretKey string, lg *zap.Logger) *BinanceUS {
	lg = lg.Named("BinanceUS")

	b := &BinanceUS{}

	b.Client = NewBinanceClient(urls.USAPIURL, apiKey, secretKey, lg)
	b.canceller = NewBinanceOrderCanceller(b.Client)
	b.orderGetter = &OrderGetter{b.Client}
	b.orderPlacer = NewOrderPlacer(b.Client)
	b.positionGetter = &PositionGetter{b.Client}
	b.urls = urls
	b.lg = lg
	return b
}

func (b *BinanceUS) RoundPrice(_ context.Context, symbol string, price *apd.Decimal, tickSize *string) (*apd.Decimal, error) {
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

func (b *BinanceUS) RoundQuantity(_ context.Context, symbol string, qty *apd.Decimal) (*apd.Decimal, error) {
	// TODO: handle this more accurately at bot side
	str := qty.Text('f')
	substr := binanceQtyFloorRE.FindString(str)
	if substr == "" {
		return nil, errors.Errorf("invalid quantity %v", qty)
	}
	return utils.FromStringErr(str)
}

func (b *BinanceUS) GetPrefix() string {
	return BINANCE_PREFIX
}

func (b *BinanceUS) GetName() string {
	return "Binance US"
}

func (b *BinanceUS) PlaceBuyOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrder(ctx, binanceSymbol, price, quantity, prefferedID, api.SideTypeBuy)
}

func (b *BinanceUS) PlaceSellOrder(ctx context.Context,
	_ bool, symbol string, price, quantity *apd.Decimal, prefferedID string,
) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrder(ctx, binanceSymbol, price, quantity, prefferedID, api.SideTypeSell)
}

func (b *BinanceUS) CancelOrder(ctx context.Context, symbol, id string) error {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.canceller.CancelOrder(ctx, binanceSymbol, id)
}

func (b *BinanceUS) ReleaseOrder(_ context.Context, symbol, id string) error {
	return nil
}

func (b *BinanceUS) GetOrderInfo(ctx context.Context, symbol, id string, _ *time.Time) (exchanges.OrderInfo, error) {
	return b.GetOrderInfoByClientOrderID(ctx, symbol, id, nil)
}

func (b *BinanceUS) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string, _ *time.Time) (exchanges.OrderInfo, error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderGetter.GetOrderInfoByClientOrderID(ctx, binanceSymbol, clientOrderID)
}

func (b *BinanceUS) GetOpenOrders(ctx context.Context) ([]exchanges.OrderDetailInfo, error) {
	return b.orderGetter.GetOpenOrders(ctx)
}

func (b *BinanceUS) GetOrders(ctx context.Context, filter exchanges.OrderFilter) (res []exchanges.OrderDetailInfo, err error) {
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

func (b *BinanceUS) GetAccount(ctx context.Context) (exchanges.Account, error) {
	return b.positionGetter.GetAccountBalances(ctx)
}

func (b *BinanceUS) GetPrice(ctx context.Context, symbol string) (*apd.Decimal, error) {
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

func (b *BinanceUS) GetTradableSymbols(ctx context.Context) ([]exchanges.SymbolInfo, error) {
	info, err := b.Client.NewExchangeInfoService().Do(ctx)
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

func (b *BinanceUS) WatchOrdersStatuses(ctx context.Context) (<-chan exchanges.OrderEvent, error) {
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
			b.lg.Sugar().Debugf("Update keep alive listen key websocket...")
			err := b.Client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
			if err != nil {
				b.lg.Sugar().Errorf("Error to keep alive user stream websocket.. wait until next ticker... reason..", err)
			}
		}
	}()

	wsEndpoint := b.urls.WSUSUserDataURL(listenKey)
	return SubscribeToOrdersV2(ctx, wsEndpoint, b.lg)
}

func (b *BinanceUS) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan exchanges.PriceEvent, error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return SubscribeToPriceV2(ctx, b.urls.WSUSAllMiniMarketsStatURL(), binanceSymbol, b.lg)
}

// PlaceBuyOrderV2 Place Buy Order with OrderType param
func (b *BinanceUS) PlaceBuyOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrderV2(ctx, binanceSymbol, price, qty, preferredID, api.SideTypeBuy, api.OrderType(orderType))
}

// PlaceSellOrderV2 Place Sell Order with OrderType param
func (b *BinanceUS) PlaceSellOrderV2(ctx context.Context, _ bool, symbol string, price, qty *apd.Decimal, preferredID string, orderType string) (id string, e error) {
	binanceSymbol := ToBinanceSymbol(symbol)
	return b.orderPlacer.PlaceOrderV2(ctx, binanceSymbol, price, qty, preferredID, api.SideTypeSell, api.OrderType(orderType))
}

func (b *BinanceUS) WatchAccountPositions(ctx context.Context) (<-chan exchanges.PositionEvent, error) {
	listenKey, err := b.Client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't take listen key")
	}

	wsEndpoint := b.urls.WSUSUserDataURL(listenKey)

	return SubscribeToPositions(ctx, wsEndpoint, b.lg)
}

func (b *BinanceUS) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
	return utils.GenClientOrderID(identifierID)
}

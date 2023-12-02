package phemex_contract

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var PhemexOrderClearedError = errors.New("phemex order was cleared")

const PHEMEX_PREFIX = "PHEMEX-"

func ToPhemexSymbol(symbol string) string {
	return strings.TrimPrefix(symbol, PHEMEX_PREFIX)
}

func ToFullSymbol(phemexSymbol string) string {
	return PHEMEX_PREFIX + phemexSymbol
}

func ConvertPhemexPriceToPriceEp(phemexSymbol string, price *apd.Decimal) (priceEp int64, scale SymbolScale, e error) {
	symbolScales, err := ScalesSubscriberInstance.GetLastSymbolScales(phemexSymbol)
	if err != nil {
		return 0, SymbolScale{}, errors.Wrap(err, "get scales error")
	}

	priceEpAPD := utils.Round(utils.Mul(price, symbolScales.PriceScaleDivider))
	result, err := priceEpAPD.Int64()
	return result, symbolScales, errors.Wrap(err, "PriceEp cannot be represented with int64")
}

type PhemexContract struct {
	client          *phemex.Client
	forkClient      *krisa_phemex_fork.Client
	ordersFetcher   *CombinedOrdersFetcher
	positionFetcher *PositionFetcher
	orderCanceller  *OrderCanceller
	orderPlacer     *OrderPlacer

	lim *PhemexRateLimiter
	lg  *zap.Logger
}

var _ exchanges.Exchange = (*PhemexContract)(nil)

func NewPhemexContract(apiKey, secretKey string, lim *PhemexRateLimiter, lg *zap.Logger) *PhemexContract {
	lg = lg.Named("PhemexContract")

	client := phemex.NewClient(apiKey, secretKey)
	forkClient := krisa_phemex_fork.NewClient(apiKey, secretKey, lg)
	cof := NewCombinedOrdersFetcher(client, forkClient, lim, lg)
	positionFetcher := NewPositionFetcher(client)

	return &PhemexContract{
		client:     client,
		forkClient: forkClient,
		// ordersFetcher:  NewOrdersFetcher(apiKey, secretKey),
		ordersFetcher:   cof,
		positionFetcher: positionFetcher,
		orderCanceller:  NewOrderCanceller(forkClient, cof, lim, lg),
		orderPlacer:     NewOrderPlacer(forkClient, cof, lim, lg),

		lim: lim,
		lg:  lg,
	}
}

// StartBackgroundJob Only one simultaneous job is allowed
func (pc *PhemexContract) StartBackgroundJob(ctx context.Context) error {
	return pc.ordersFetcher.Start(ctx)
}

func (pc *PhemexContract) GetPrefix() string {
	return PHEMEX_PREFIX
}

func (pc *PhemexContract) GetName() string {
	return "Phemex Contract"
}

func (pc *PhemexContract) RoundQuantity(_ context.Context, symbol string, quantity *apd.Decimal) (*apd.Decimal, error) {
	// FIXME: add good rounding
	qtyRounded := utils.Round(quantity)
	if utils.Eq(qtyRounded, utils.Zero) {
		qtyRounded = utils.One
	}

	_, err := utils.ToIntegerInFloat64(qtyRounded)
	if err != nil {
		return nil, err
	}

	return qtyRounded, nil
}

func (pc *PhemexContract) RoundPrice(_ context.Context, symbol string, price *apd.Decimal, tickSize *string) (*apd.Decimal, error) {
	symbol = ToPhemexSymbol(symbol)

	priceEp, scale, err := ConvertPhemexPriceToPriceEp(symbol, price)
	if err != nil {
		return nil, err
	}
	priceResult := utils.Div(apd.New(priceEp, 0), scale.PriceScaleDivider)

	return priceResult, nil
}

// PlaceBuyOrder This method should use `clientOrderID` if it's possible
func (pc *PhemexContract) PlaceBuyOrder(ctx context.Context,
	isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string,
) (id string, e error) {
	symbol = ToPhemexSymbol(symbol)
	return pc.orderPlacer.PlaceOrder(ctx, isRetry, symbol, price, qty, clientOrderID, krisa_phemex_fork.SideTypeBuy)
}

// PlaceSellOrder This method should use `clientOrderID` if it's possible
func (pc *PhemexContract) PlaceSellOrder(ctx context.Context,
	isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string,
) (id string, e error) {
	symbol = ToPhemexSymbol(symbol)
	return pc.orderPlacer.PlaceOrder(ctx, isRetry, symbol, price, qty, clientOrderID, krisa_phemex_fork.SideTypeSell)
}

// CancelOrder can return `OrderExecutedError` in case of executed order.
// can return `OrderNotFoundError` in case of not found order.
func (pc *PhemexContract) CancelOrder(ctx context.Context, symbol, id string) error {
	// ?OPTIMIZATION: in case of order changing there is AmendOrder/ReplaceOrder method
	symbol = ToPhemexSymbol(symbol)
	// ?OPTIMIZATION: for first cancelation no pre check is required
	return pc.orderCanceller.CancelOrder(ctx, symbol, id)
}

func (pc *PhemexContract) ReleaseOrder(_ context.Context, symbol, id string) error {
	return nil
}

func (pc *PhemexContract) checkOrderDate(createdAt *time.Time) error {
	if createdAt != nil {
		// We can have either 60 days or 2 months. After that orders will removed from exchange.
		// I (Mikhail) don't know Phemex rules: if we will subtract 60 days from
		// 31 March we can jump to 31 Jan and this is more than 2 months.
		// So I've selected 55 days as a best approach

		twoMonthsAgo := time.Now().AddDate(0, 0, -55)
		if createdAt.Before(twoMonthsAgo) {
			return errors.New("order is too old and can't be used")
		}
	}
	return nil
}

func (pc *PhemexContract) GetOrderInfo(ctx context.Context, symbol, id string, createdAt *time.Time) (exchanges.OrderInfo, error) {
	if err := pc.checkOrderDate(createdAt); err != nil {
		return exchanges.OrderInfo{}, err
	}

	symbol = ToPhemexSymbol(symbol)
	oInfo, _, err := pc.ordersFetcher.GetOrderInfoByOrderID(ctx, symbol, id)
	return oInfo, err
}

func (pc *PhemexContract) GetOrderInfoByClientOrderID(ctx context.Context,
	symbol, clientOrderID string, createdAt *time.Time,
) (exchanges.OrderInfo, error) {
	if err := pc.checkOrderDate(createdAt); err != nil {
		return exchanges.OrderInfo{}, err
	}

	symbol = ToPhemexSymbol(symbol)
	oInfo, _, err := pc.ordersFetcher.GetOrderInfoByClientOrderID(ctx, symbol, clientOrderID)
	return oInfo, err
}

func (pc *PhemexContract) GetOpenOrders(ctx context.Context) ([]exchanges.OrderDetailInfo, error) {
	return pc.ordersFetcher.GetOpenOrders(ctx)
}

func (pc *PhemexContract) GetOrders(ctx context.Context, filter exchanges.OrderFilter) (res []exchanges.OrderDetailInfo, err error) {
	if filter.Symbol == nil {
		return res, errors.New("symbol is empty!")
	}
	symbol := ToPhemexSymbol(*filter.Symbol)
	return pc.ordersFetcher.GetHistoryOrders(
		ctx,
		symbol,
		filter.OrderID,
		filter.ClientOrderID,
	)
}

func (pc *PhemexContract) GetAccount(ctx context.Context) (exchanges.Account, error) {
	return pc.positionFetcher.GetAccountPosition(ctx)
}

func (pc *PhemexContract) GetTradableSymbols(ctx context.Context) ([]exchanges.SymbolInfo, error) {
	data, err := pc.forkClient.NewProductsService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to make request")
	}
	result := []exchanges.SymbolInfo{}
	for _, product := range data.Products {
		if product.Type != krisa_phemex_fork.PerpetualProductType {
			continue
		}
		if product.Status != krisa_phemex_fork.ListedProductStatus {
			continue
		}
		filters := []map[string]interface{}{
			{
				"contractSize": product.ContractSize,
				"lotSize":      product.LotSize,
				"tickSize":     product.TickSize,
			},
		}

		fullSymbol := ToFullSymbol(product.Symbol)
		displaySymbol := PHEMEX_PREFIX + strings.ReplaceAll(product.DisplaySymbol, " / ", "")
		displayName := fmt.Sprintf("%s (%s-Margin)", displaySymbol, product.SettleCurrency)
		result = append(result, exchanges.SymbolInfo{
			DisplayName:    displayName,
			Symbol:         fullSymbol,
			OriginalSymbol: product.Symbol,
			Filters:        filters,
		})
	}
	return result, nil
}

// WatchOrdersStatuses Returns control immediately
// If error is sent then channel will be closed automatically
// Channel will be closed in two ways: by context and by disconnection
func (pc *PhemexContract) WatchOrdersStatuses(ctx context.Context) (<-chan exchanges.OrderEvent, error) {
	in, err := SubscribeToOrders(ctx, pc.client, pc.lg)
	if err != nil {
		return nil, err
	}

	out := make(chan exchanges.OrderEvent, 100)
	go func() {
		defer close(out)
		for ev := range in {
			out <- ev.event
		}
	}()
	return out, nil
}

// WatchSymbolPrice Returns control immediately
// If error is sent then channel will be closed automatically
// Channel will be closed in two ways: by context and by disconnection
func (pc *PhemexContract) WatchSymbolPrice(ctx context.Context, symbol string) (<-chan exchanges.PriceEvent, error) {
	phemexSymbol := ToPhemexSymbol(symbol)
	return SubscribeToPrice(ctx, phemexSymbol, pc.lg)
}

// This method use to PlaceOrder with Order Type as Parameter
//func (pc *PhemexContract) PlaceBuyOrderV2(ctx context.Context,
//	isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string) (id string, e error) {
//	symbol = ToPhemexSymbol(symbol)
//	orderTypePlacing := ToOrderType(orderType)
//	return pc.orderPlacer.PlaceOrderV2(ctx, isRetry, symbol, price, qty, clientOrderID, krisa_phemex_fork.SideTypeBuy, orderTypePlacing)
//}

func ToOrderType(orderType string) krisa_phemex_fork.OrderType {
	return krisa_phemex_fork.OrderType(orderType)
}

// PlaceBuyOrderV2 This method use to PlaceOrder with Order Type as Parameter
func (pc *PhemexContract) PlaceBuyOrderV2(ctx context.Context,
	isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string,
) (id string, e error) {
	symbol = ToPhemexSymbol(symbol)
	orderTypePlacing := ToOrderType(orderType)
	return pc.orderPlacer.PlaceOrderV2(ctx, isRetry, symbol, price, qty, clientOrderID, krisa_phemex_fork.SideTypeBuy, orderTypePlacing)
}

// PlaceSellOrderV2 This method use to PlaceOrder with OrderType as Parameter
func (pc *PhemexContract) PlaceSellOrderV2(ctx context.Context,
	isRetry bool, symbol string, price, qty *apd.Decimal, clientOrderID string, orderType string,
) (id string, e error) {
	symbol = ToPhemexSymbol(symbol)
	orderTypePlacing := ToOrderType(orderType)
	return pc.orderPlacer.PlaceOrderV2(ctx, isRetry, symbol, price, qty, clientOrderID, krisa_phemex_fork.SideTypeSell, orderTypePlacing)
}

// WatchAccountPositions Returns control immediately
// If error is sent then channel will be closed automatically
// Channel will be closed in two ways: by context and by disconnection
func (pc *PhemexContract) WatchAccountPositions(ctx context.Context) (<-chan exchanges.PositionEvent, error) {
	in, err := SubscribeToPositions(ctx, pc.client, pc.lg)
	if err != nil {
		return nil, err
	}

	out := make(chan exchanges.PositionEvent, 100)
	go func() {
		defer close(out)
		for ev := range in {
			out <- ev.event
		}
	}()
	return out, nil
}

func (b *PhemexContract) GenerateClientOrderID(ctx context.Context, identifierID string) (string, error) {
	return utils.GenClientOrderID(identifierID)
}

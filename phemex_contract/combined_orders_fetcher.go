package phemex_contract

import (
	"context"
	"sync"
	"time"

	"github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var orderNotFoundInCacheError error = errors.New("order not found in cache")

type SymbolClientOrderID struct {
	Symbol, ClientOrderID string
}

type SymbolOrderID struct {
	Symbol, OrderID string
}

func InsertToOrdersFetcherMultimap(m map[SymbolClientOrderID]map[string]struct{}, key SymbolClientOrderID, value string) {
	values, ok := m[key]
	if !ok {
		values = map[string]struct{}{}
		m[key] = values
	}
	values[value] = struct{}{}
}

func TakeFirstValueOfStringSet(s map[string]struct{}) string {
	for x := range s {
		return x
	}
	return ""
}

type orderResponse struct {
	fields orderFields
	status exchanges.OrderStatusType
}

func phemexOrderResponseToOrderResponse(oi exchanges.OrderInfo, or *krisa_phemex_fork.OrderResponse) *orderResponse {
	return &orderResponse{
		status: oi.Status,
		fields: orderFields{
			ClOrdID:     or.ClOrdID,
			OrdType:     or.OrderType,
			OrderQty:    or.OrderQty,
			PriceEp:     or.PriceEp,
			Side:        or.Side,
			Symbol:      or.Symbol,
			TimeInForce: or.TimeInForce,
		},
	}
}

type WSOrdersCache struct {
	clientOrderIDToOrderIDs map[SymbolClientOrderID]map[string]struct{} // TODO: GC
	orderIDToLastState      map[SymbolOrderID]*orderResponse            // TODO: GC
}

func NewWSOrdersCache() *WSOrdersCache {
	return &WSOrdersCache{
		clientOrderIDToOrderIDs: map[SymbolClientOrderID]map[string]struct{}{},
		orderIDToLastState:      map[SymbolOrderID]*orderResponse{},
	}
}

func (c *WSOrdersCache) InsertOrder(
	orderIDSelector SymbolOrderID,
	clientOrderIDSelector SymbolClientOrderID,
	or *orderResponse,
) {
	InsertToOrdersFetcherMultimap(c.clientOrderIDToOrderIDs, clientOrderIDSelector, orderIDSelector.OrderID)
	c.orderIDToLastState[orderIDSelector] = or
}

func (c *WSOrdersCache) GetByOrderID(orderIDSelector SymbolOrderID) *orderResponse {
	return c.orderIDToLastState[orderIDSelector]
}

func (to *WSOrdersCache) CopyFrom(from *WSOrdersCache) {
	for selector, fromValues := range from.clientOrderIDToOrderIDs {
		toValues, ok := to.clientOrderIDToOrderIDs[selector]
		if !ok {
			toValues = map[string]struct{}{}
			to.clientOrderIDToOrderIDs[selector] = toValues
		}

		for value := range fromValues {
			toValues[value] = struct{}{}
		}
	}

	for selector, state := range from.orderIDToLastState {
		to.orderIDToLastState[selector] = state
	}
}

type CombinedOrdersFetcher struct {
	client     *phemex.Client
	forkClient *krisa_phemex_fork.Client

	mutex                sync.RWMutex
	subscribedToOrdersAt time.Time

	finalOrdersCache *WSOrdersCache
	allOrdersCache   *WSOrdersCache

	lim *PhemexRateLimiter

	lg *zap.Logger
}

func NewCombinedOrdersFetcher(client *phemex.Client, forkClient *krisa_phemex_fork.Client, lim *PhemexRateLimiter, logger *zap.Logger) *CombinedOrdersFetcher {
	return &CombinedOrdersFetcher{
		client:     client,
		forkClient: forkClient,

		finalOrdersCache: NewWSOrdersCache(),
		allOrdersCache:   NewWSOrdersCache(),

		lim: lim,
		lg:  logger.Named("CombinedOrdersFetcher"),
	}
}

func (cof *CombinedOrdersFetcher) subscribeWithAutoreconnect(ctx context.Context) (<-chan WSOrderEvent, error) {
	chanShifter := func(in <-chan WSOrderEvent, out chan<- WSOrderEvent) {
		for ev := range in {
			out <- ev
		}
	}

	in, err := SubscribeToOrders(ctx, cof.client, cof.lg)
	if err != nil {
		return nil, errors.Wrap(err, "can't subscribe")
	}

	intermediate := make(chan WSOrderEvent, 100)
	out := make(chan WSOrderEvent, 100)
	go chanShifter(in, intermediate)

	reconnectAttempt := atomic.NewUint32(0)
	reconnect := func() {
		reconnectAttempt.Inc()
		serialAttempt := 1

		cof.lg.Info("Reconnecting...", zap.Uint32("reconnectAttempt", reconnectAttempt.Load()))
		for {
			defer func() { serialAttempt++ }()
			lg := cof.lg.With(
				zap.Uint32("attempt", reconnectAttempt.Load()),
				zap.Int("serialAttempt", serialAttempt))

			in, err = SubscribeToOrders(ctx, cof.client, cof.lg)
			if err == nil {
				lg.Info("Reconnected successfully")
				out <- WSOrderEvent{event: exchanges.OrderEvent{Reconnected: &struct{}{}}}
				go chanShifter(in, intermediate)
				return
			}

			lg.Warn("Reconnect error", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			default:
			}
			time.Sleep(time.Second * 30)
		}
	}

	go func() {
		clearCache := func() {
			cof.mutex.Lock()
			// ? OPTIMIZATION: instead of copying we can use double layer
			// ? cache with "final" and "non-final" caches on the leafs
			cof.allOrdersCache = NewWSOrdersCache()
			cof.allOrdersCache.CopyFrom(cof.finalOrdersCache)
			cof.mutex.Unlock()
		}

		defer close(out)
		for {
			select {
			case <-ctx.Done():
				clearCache()
				return
			case ev := <-intermediate:
				switch {
				case ev.event.DisconnectedWithErr != nil:
					cof.lg.Warn("Disconnected with error", zap.Error(ev.event.DisconnectedWithErr))
					clearCache()
					reconnect()
				case ev.event.Reconnected != nil:
					out <- ev
				case ev.event.Payload != nil:
					out <- ev
				default:
					cof.lg.Warn("Incorrect event", zap.Any("ev", ev))
				}
			}
		}
	}()
	return out, nil
}

func (cof *CombinedOrdersFetcher) Start(ctx context.Context) error {
	events, err := cof.subscribeWithAutoreconnect(ctx)
	if err != nil {
		return errors.Wrap(err, "can't subscribe")
	}
	cof.lg.Info("Subscribed")

	cof.mutex.Lock()
	cof.subscribedToOrdersAt = time.Now()
	cof.mutex.Unlock()

	go func() {
		for {
			var ev WSOrderEvent
			select {
			case <-ctx.Done():
				return
			case ev = <-events:
			}

			switch {
			case ev.event.DisconnectedWithErr != nil:
				cof.lg.Warn("Phemex Combined Orders Fetcher watcher disconnected", zap.Error(ev.event.DisconnectedWithErr))
			case ev.event.Reconnected != nil:
				cof.lg.Info("Reconnected")
			case ev.event.Payload != nil:
				if ev.fields == nil {
					cof.lg.Error("ev.fields is empty while ev.event.Payload is not")
					continue
				}

				orderID := ev.event.Payload.OrderID

				orderIDSelector := SymbolOrderID{
					Symbol:  ev.fields.Symbol,
					OrderID: orderID,
				}

				clientOrderIDSelector := SymbolClientOrderID{
					Symbol:        ev.fields.Symbol,
					ClientOrderID: ev.fields.ClOrdID,
				}

				cof.mutex.Lock()
				{
					storedState := cof.finalOrdersCache.GetByOrderID(orderIDSelector)
					if storedState != nil && storedState.status.IsFinalStatus() {
						// After Canceled status can follow Rejected status which should be ignored
						cof.mutex.Unlock()
						continue
					}

					or := &orderResponse{fields: *ev.fields, status: ev.event.Payload.OrderStatus}

					cof.allOrdersCache.InsertOrder(orderIDSelector, clientOrderIDSelector, or)

					if ev.event.Payload.OrderStatus.IsFinalStatus() {
						cof.finalOrdersCache.InsertOrder(orderIDSelector, clientOrderIDSelector, or)
					}
				}
				cof.mutex.Unlock()
			default:
				cof.lg.Warn("Incorrect event", zap.Any("ev", ev))
			}
		}
	}()
	return nil
}

func (cof *CombinedOrdersFetcher) unifiedGetOrderInfo(
	ctx context.Context, svc *krisa_phemex_fork.QueryOrderService,
) (exchanges.OrderInfo, *orderResponse, error) {
	cof.lim.Other.Lim.Wait()
	resp, rateLimHeaders, err := svc.Do(ctx)
	cof.lim.Apply(rateLimHeaders)
	if IsAPINotFoundError(err) {
		return exchanges.OrderInfo{}, nil, errors.Wrap(err, "unexpected not found error")
	}
	if err != nil {
		return exchanges.OrderInfo{}, nil, errors.Wrap(err, "unable to send request")
	}

	if len(resp) == 0 {
		return exchanges.OrderInfo{}, nil, exchanges.OrderNotFoundError
	}
	if len(resp) > 1 {
		return exchanges.OrderInfo{}, nil, errors.Errorf("too many orders in response")
	}

	oInfo, err := convertOrder(resp[0])
	if err != nil {
		return exchanges.OrderInfo{}, nil, errors.Wrap(err, "can't convert order")
	}
	return oInfo, phemexOrderResponseToOrderResponse(oInfo, resp[0]), nil
}

func (cof *CombinedOrdersFetcher) GetOrderInfoByClientOrderID(
	ctx context.Context, symbol, clientOrderID string,
) (exchanges.OrderInfo, *orderResponse, error) {
	oi, or, err := cof.getOrderFromAOPStreamDataByClientOrderID(cof.allOrdersCache, symbol, clientOrderID, false)
	switch {
	case err == nil:
		return oi, or, nil
	case errors.Is(err, orderNotFoundInCacheError):
		// cache miss
	case err != nil:
		return oi, or, errors.Wrap(err, "can't get order from WS")
	}

	oi, or, err = cof.unifiedGetOrderInfo(
		ctx,
		cof.forkClient.NewQueryOrderService().
			Symbol(symbol).
			ClOrderID(clientOrderID))

	if errors.Is(err, PhemexOrderClearedError) {
		oi, or, err := cof.getOrderFromAOPStreamDataByClientOrderID(
			cof.finalOrdersCache, symbol, clientOrderID, true)
		if err != nil {
			return oi, or, errors.Wrap(err, "fallback error")
		}
		return oi, or, nil
	}

	return oi, or, err
}

func (cof *CombinedOrdersFetcher) GetOrderInfoByOrderID(
	ctx context.Context, symbol, id string,
) (exchanges.OrderInfo, *orderResponse, error) {
	oi, or, err := cof.getOrderFromAOPStreamDataByOrderID(cof.allOrdersCache, symbol, id, false, true)
	switch {
	case err == nil:
		return oi, or, nil
	case errors.Is(err, orderNotFoundInCacheError):
		// cache miss
	case err != nil:
		return oi, or, errors.Wrap(err, "can't get order from WS")
	}

	oi, or, err = cof.unifiedGetOrderInfo(
		ctx,
		cof.forkClient.NewQueryOrderService().
			Symbol(symbol).
			OrderID(id))

	if errors.Is(err, PhemexOrderClearedError) {
		oi, or, err := cof.getOrderFromAOPStreamDataByOrderID(
			cof.finalOrdersCache, symbol, id, true, true)
		if err != nil {
			return oi, or, errors.Wrap(err, "fallback error")
		}
		return oi, or, nil
	}

	return oi, or, err
}

func (cof *CombinedOrdersFetcher) buildOrderDetailInfo(order *phemex.OrderResponse) (res exchanges.OrderDetailInfo, err error) {
	price := utils.FromFloat64(order.Price)

	quantity := utils.FromFloat64(order.OrderQty)

	executedQty := utils.FromFloat64(order.ClosedSize)

	orderStatusType, err := convertOrderStatus(order.OrdStatus)
	if err != nil {
		err = errors.Wrap(err, "convert order status")
		return res, err
	}

	orderType := mapOrderType(string(order.OrderType))

	orderSide := mapOrderSide(string(order.Side))

	orderTimeInForce := mapOrderTimeInForce(string(order.TimeInForce))

	res = exchanges.OrderDetailInfo{
		Symbol:        order.Symbol,
		ID:            order.OrderID,
		ClientOrderID: &order.ClOrdID,
		Price:         price,
		Quantity:      quantity,
		ExecutedQty:   executedQty,
		Status:        orderStatusType,
		OrderType:     orderType,
		Time:          order.ActionTimeNs,
		OrderSide:     orderSide,
		TimeInForce:   orderTimeInForce,
		StopPrice:     nil,
		QuoteQuantity: nil,
	}

	return
}

func (cof *CombinedOrdersFetcher) GetOpenOrders(ctx context.Context) (res []exchanges.OrderDetailInfo, err error) {
	openOrders, err := cof.client.NewListOpenOrdersService().Do(ctx)
	if err != nil {
		err = errors.Wrap(err, "error phemex get open orders")
		return
	}

	for _, openOrder := range openOrders {
		orderDetailInfo, err := cof.buildOrderDetailInfo(openOrder)
		if err != nil {
			return res, err
		}
		res = append(res, orderDetailInfo)
	}
	return
}

func (cof *CombinedOrdersFetcher) GetHistoryOrders(
	ctx context.Context,
	symbol string,
	orderID *string,
	clientOrderID *string,
) (res []exchanges.OrderDetailInfo, err error) {
	orderClient := cof.client.NewQueryOrderService().Symbol(symbol)

	if orderID != nil {
		orderClient.OrderID(*orderID)
	}

	if clientOrderID != nil {
		orderClient.ClOrderID(*clientOrderID)
	}

	orders, err := orderClient.Do(ctx)
	if err != nil {
		err = errors.Wrap(err, "error phemex get open orders")
		return
	}

	for _, order := range orders {
		orderDetailInfo, err := cof.buildOrderDetailInfo(order)
		if err != nil {
			return res, err
		}
		res = append(res, orderDetailInfo)
	}
	return
}

// thread safe
func (cof *CombinedOrdersFetcher) fallbackDelay(needToLock bool) {
	if needToLock {
		cof.mutex.RLock()
	}
	subscribedAt := cof.subscribedToOrdersAt
	if needToLock {
		cof.mutex.RUnlock()
	}

	delay := 10 * time.Second // Instead of 10 seconds it's better to wait for first WS message from Phemex
	if subscribedAt == (time.Time{}) {
		time.Sleep(delay)
	} else {
		elapsed := time.Since(subscribedAt)
		if elapsed < delay {
			time.Sleep(delay - elapsed)
		}
	}
}

func (cof *CombinedOrdersFetcher) getOrderFromAOPStreamDataByClientOrderID(
	cache *WSOrdersCache, symbol, clientOrderID string, needToWait bool,
) (exchanges.OrderInfo, *orderResponse, error) {
	if needToWait {
		cof.fallbackDelay(true)
	}

	cof.mutex.RLock()
	defer cof.mutex.RUnlock()

	orderIDs := cache.clientOrderIDToOrderIDs[SymbolClientOrderID{
		Symbol: symbol, ClientOrderID: clientOrderID}]

	switch {
	case len(orderIDs) > 1:
		return exchanges.OrderInfo{}, nil, errors.New("too many orders with same `clientOrderID`")
	case len(orderIDs) == 0:
		return exchanges.OrderInfo{}, nil, orderNotFoundInCacheError
	}

	orderID := TakeFirstValueOfStringSet(orderIDs)
	oi, or, err := cof.getOrderFromAOPStreamDataByOrderID(cache, symbol, orderID, false, false)
	return oi, or, errors.Wrap(err, "can't take WS order by order ID")
}

func (cof *CombinedOrdersFetcher) getOrderFromAOPStreamDataByOrderID(
	cache *WSOrdersCache, symbol, orderID string, needToWait, needToLock bool,
) (exchanges.OrderInfo, *orderResponse, error) {
	if needToWait {
		cof.fallbackDelay(needToLock)
	}

	if needToLock {
		cof.mutex.Lock()
		defer cof.mutex.Unlock()
	}

	state := cache.GetByOrderID(SymbolOrderID{
		Symbol:  symbol,
		OrderID: orderID,
	})
	if state != nil {
		return exchanges.OrderInfo{
			ID:            orderID,
			ClientOrderID: &state.fields.ClOrdID,
			Status:        state.status,
		}, state, nil
	}

	return exchanges.OrderInfo{}, nil, orderNotFoundInCacheError
}

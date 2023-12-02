package binance

import (
	"context"
	"encoding/json"
	"time"

	api "github.com/adshao/go-binance/v2"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance/adshao_binance"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const orderUpdateEventType string = "executionReport"
const orderUpdateFuturesEventType string = "ORDER_TRADE_UPDATE"

type userDataStreamCommonMessage struct {
	EventType string `json:"e"` // "e": "executionReport", // Event type

	// !Don't remove this field. It is required due to case-insensitive stdlib JSON
	EventTime int64 `json:"E"` // "E": 1499405658658,     // Event time
}

type OrderUpdateFutures struct {
	EventType string                 `json:"e"` // "e": "ORDER_TRADE_UPDATE", // Event type
	EventTime int64                  `json:"E"` // "E": 1499405658658,     // Event time
	OrderData OrderUpdateDataFutures `json:"o"`
}

type OrderUpdateDataFutures struct {
	ClientOrderID string `json:"c"` // "c": "mUvoqJxFIILMdfAW5iGSOW", // Client order ID              // Original client order ID; This is the ID of the order being canceled

	CurrentExecutionType string `json:"x"` // "x": "NEW", // Current execution type
	CurrentOrderStatus   string `json:"X"` // "X": "NEW", // Current order status

	Symbol string `json:"s"` // "s": "ETHBTC", // Symbol
	Side   string `json:"S"` // "S": "BUY",    // Side

}

type OrderUpdate struct {
	// !WARNING Keep upper and lower case letters -- this is workaround to make case-sensetive JSON
	EventType string `json:"e"` // "e": "executionReport", // Event type
	EventTime int64  `json:"E"` // "E": 1499405658658,     // Event time

	ClientOrderID         string  `json:"c"` // "c": "mUvoqJxFIILMdfAW5iGSOW", // Client order ID
	OriginalClientOrderID *string `json:"C"` // "C": null,                     // Original client order ID; This is the ID of the order being canceled

	CurrentExecutionType string `json:"x"` // "x": "NEW", // Current execution type
	CurrentOrderStatus   string `json:"X"` // "X": "NEW", // Current order status

	Symbol string `json:"s"` // "s": "ETHBTC", // Symbol
	Side   string `json:"S"` // "S": "BUY",    // Side

	// EventType                string  `json:"e"` // "e": "executionReport",        // Event type
	// EventTime                int64   `json:"E"` // "E": 1499405658658,            // Event time
	// ClientOrderID string `json:"c"` // "c": "mUvoqJxFIILMdfAW5iGSOW",             // Client order ID
	// OrderType                string  `json:"o"` // "o": "LIMIT",                  // Order type
	// TimeInForce              string  `json:"f"` // "f": "GTC",                    // Time in force
	// OrderQuantity            string  `json:"q"` // "q": "1.00000000",             // Order quantity
	// OrderPrice               string  `json:"p"` // "p": "0.10264410",             // Order price
	// StopPrice                string  `json:"P"` // "P": "0.00000000",             // Stop price
	// IcebergQuantity          string  `json:"F"` // "F": "0.00000000",             // Iceberg quantity
	// OrderListID              int64   `json:"g"` // "g": -1,                       // OrderListId
	// OriginalClientOrderID *string `json:"C"` // "C": null,                        // Original client order ID; This is the ID of the order being canceled
	// CurrentExecutionType     string  `json:"x"` // "x": "NEW",                    // Current execution type
	// CurrentOrderStatus string `json:"X"` // "X": "NEW",                           // Current order status
	// OrderRejectReason        string  `json:"r"` // "r": "NONE",                   // Order reject reason; will be an error code.
	// OrderID                  int64   `json:"i"` // "i": 4293153,                  // Order ID
	// LastExecutedQuantity     string  `json:"l"` // "l": "0.00000000",             // Last executed quantity
	// CumulativeFilledQuantity string  `json:"z"` // "z": "0.00000000",             // Cumulative filled quantity
	// LastExecutedPrice        string  `json:"L"` // "L": "0.00000000",             // Last executed price
	// CommissionAmount         string  `json:"n"` // "n": "0",                      // Commission amount
	// CommissionAsset          *string `json:"N"` // "N": null,                     // Commission asset
	// TransactionTime          int64   `json:"T"` // "T": 1499405658657,            // Transaction time
	// TradeID                  int64   `json:"t"` // "t": -1,                       // Trade ID
	// Ignore ?             `json:"I"` // "I": 8641984,                  // Ignore
	// IsTheOrderOnTheBook     bool `json:"w"` // "w": true,                     // Is the order on the book?
	// IsThisTradeTheMakerSide bool `json:"m"` // "m": false,                    // Is this trade the maker side?
	// Ignore ?             `json:"M"` // "M": false,                    // Ignore
	// OrderCreationTime                      int64  `json:"O"` // "O": 1499405658657,            // Order creation time
	// CumulativeQuoteAssetTransactedQuantity string `json:"Z"` // "Z": "0.00000000",             // Cumulative quote asset transacted quantity
	// LastQuoteAssetTransactedQuantity       string `json:"Y"` // "Y": "0.00000000",             // Last quote asset transacted quantity (i.e. lastPrice * lastQty)
	// QuoteOrderQty                          string `json:"Q"` // "Q": "0.00000000"              // Quote Order Qty
}

func SubscribeToOrders(
	ctx context.Context,
	urls BinanceURLs,
	client *api.Client,
	lg *zap.Logger,
) (<-chan exchanges.OrderEvent, error) {
	listenKey, err := client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't take listen key")
	}

	// Add capability to keep alive the listen key
	ticker := time.NewTicker(20 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			err := client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
			if err != nil {
				lg.Sugar().Errorf("Error to keep alive user stream websocket.. wait until next ticker... reason..", err)
			}
		}
	}()

	cfg := adshao_binance.WSConfig{
		Endpoint:  urls.WSUserDataURL(listenKey),
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Orders"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.OrderEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			ok, err := isOrderEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't understand type of user datastream event"),
				}
				return
			}

			if !ok {
				continue
			}

			result, err := mapToOrderEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't parse OrderEvent"),
				}
				return
			}

			out <- exchanges.OrderEvent{
				Payload: result,
			}
		}
	}()

	return out, nil
}

// Similar with SubscribeToOrders but it accepts ws endpoint instead of urls object
func SubscribeToOrdersV2(
	ctx context.Context,
	wsEndpoint string,
	lg *zap.Logger,
) (<-chan exchanges.OrderEvent, error) {
	cfg := adshao_binance.WSConfig{
		Endpoint:  wsEndpoint,
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Orders"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.OrderEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			ok, err := isOrderEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't understand type of user datastream event"),
				}
				return
			}

			if !ok {
				continue
			}

			result, err := mapToOrderEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't parse OrderEvent"),
				}
				return
			}

			out <- exchanges.OrderEvent{
				Payload: result,
			}
		}
	}()

	return out, nil
}

func SubscribeToOrdersFutures(
	ctx context.Context,
	wsEndpoint string,
	lg *zap.Logger,
) (<-chan exchanges.OrderEvent, error) {
	cfg := adshao_binance.WSConfig{
		Endpoint:  wsEndpoint,
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Orders"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.OrderEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				lg.Sugar().Errorf("error while reading event: %s", msg.DisconnectedWithErr)
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			ok, err := isOrderEventFuturesPayload(msg.Payload)
			if err != nil {
				lg.Sugar().Errorf("error while checking order update event type: %s", msg.DisconnectedWithErr)

				out <- exchanges.OrderEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't understand type of user datastream event"),
				}
				return
			}

			if !ok {
				continue
			}

			result, err := mapToOrderFuturesEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.OrderEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't parse OrderEvent"),
				}
				return
			}

			out <- exchanges.OrderEvent{
				Payload: result,
			}
		}
	}()

	return out, nil
}

func mapToOrderEventPayload(message []byte) (p *exchanges.OrderEventPayload, e error) {
	orderUpdate := OrderUpdate{}
	err := json.Unmarshal(message, &orderUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal JSON")
	}

	// TODO: there is sense to add symbol to payload too.
	result := &exchanges.OrderEventPayload{}
	if orderUpdate.OriginalClientOrderID != nil && *orderUpdate.OriginalClientOrderID != "" {
		result.OrderID = *orderUpdate.OriginalClientOrderID
	} else {
		result.OrderID = orderUpdate.ClientOrderID
	}

	fullSymbol := ToFullSymbol(orderUpdate.Symbol)
	result.Symbol = &fullSymbol

	result.OrderStatus = mapOrderStatusType(orderUpdate.CurrentOrderStatus)
	return result, nil
}

func isOrderEventPayload(message []byte) (bool, error) {
	data := userDataStreamCommonMessage{}
	err := json.Unmarshal(message, &data)
	if err != nil {
		return false, errors.Wrap(err, string(message))
	}

	return data.EventType == orderUpdateEventType, nil
}

func isOrderEventFuturesPayload(message []byte) (bool, error) {
	data := userDataStreamCommonMessage{}
	err := json.Unmarshal(message, &data)
	if err != nil {
		return false, errors.Wrap(err, string(message))
	}

	return data.EventType == orderUpdateFuturesEventType, nil
}

func mapToOrderFuturesEventPayload(message []byte) (p *exchanges.OrderEventPayload, e error) {
	orderUpdate := OrderUpdateFutures{}
	err := json.Unmarshal(message, &orderUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal JSON")
	}

	result := &exchanges.OrderEventPayload{}

	result.OrderID = orderUpdate.OrderData.ClientOrderID

	fullSymbol := ToFullSymbol(orderUpdate.OrderData.Symbol)
	result.Symbol = &fullSymbol

	result.OrderStatus = mapOrderStatusType(orderUpdate.OrderData.CurrentOrderStatus)
	return result, nil
}

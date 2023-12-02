package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	api "github.com/adshao/go-binance/v2"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCancelIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestCancelIntegration")
	}

	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	ctx := context.Background()
	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())

	symbol := "LTCUSDT"
	clientOrderIDCanceled := "RUN76-b40bfc43"    // CANCELED
	clientOrderIDNotFound := "RUN76-12323123123" // NOT FOUND
	clientOrderIDFilled := "RUN166-c213443c"     // FILLED

	err := ex.CancelOrder(ctx, symbol, clientOrderIDCanceled)
	assert.NoError(t, err)

	err = ex.CancelOrder(ctx, symbol, clientOrderIDNotFound)
	assert.True(t, errors.Is(err, exchanges.OrderNotFoundError), err.Error())

	err = ex.CancelOrder(ctx, symbol, clientOrderIDFilled)
	assert.True(t, errors.Is(err, exchanges.OrderExecutedError), err.Error())
}

func TestGetOrderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestGetOrderIntegration")
	}

	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")

	ctx := context.Background()
	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())

	symbol := "LTCUSDT"
	clientOrderIDCanceled := "RUN76-b40bfc43"    // CANCELED
	clientOrderIDNotFound := "RUN76-12323123123" // NOT FOUND
	clientOrderIDFilled := "RUN166-c213443c"     // FILLED

	order, err := ex.GetOrderInfo(ctx, symbol, clientOrderIDCanceled, nil)
	assert.NoError(t, err)
	assert.Equal(t, clientOrderIDCanceled, *order.ClientOrderID)
	assert.Equal(t, clientOrderIDCanceled, order.ID)
	assert.Equal(t, exchanges.CanceledOST, order.Status)

	_, err = ex.GetOrderInfo(ctx, symbol, clientOrderIDNotFound, nil)
	assert.True(t, errors.Is(err, exchanges.OrderNotFoundError), err.Error())

	order, err = ex.GetOrderInfo(ctx, symbol, clientOrderIDFilled, nil)
	assert.NoError(t, err)
	assert.Equal(t, clientOrderIDFilled, *order.ClientOrderID)
	assert.Equal(t, clientOrderIDFilled, order.ID)
	assert.Equal(t, exchanges.FilledOST, order.Status)

	openOrder, err := ex.GetOpenOrders(ctx)
	assert.NoError(t, err)
	fmt.Println(openOrder)
}

func TestTestOrderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestOrderIntegration")
	}

	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	client := api.NewClient(key, secret)

	op := binance.NewOrderPlacer(client)

	symbol := "LTCUSDT"
	price := utils.FromString("0.01")
	quantity := utils.FromString("0.0001")
	id, err := uuid.NewV4()
	assert.NoError(t, err)

	ctx := context.Background()

	err = op.TestOrder(ctx, symbol, price, quantity, id.String(), api.SideTypeBuy)
	assert.True(t, errors.Is(err, exchanges.NewOrderRejectedError), err.Error())
}

func TestPriceSubscriberIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestPriceSubscriberIntegration")
	}

	key, secret := "", ""
	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := ex.WatchSymbolPrice(ctx, "LTCUSDT")
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		ev := <-ch
		return ev.Payload != nil && ev.Payload.Cmp(utils.Zero) > 0
	}, 10*time.Second, time.Second/10)
}

// go test ./... -run TestOrdersSubscriberIntegration -v -timeout 30000ms
func xTestOrdersSubscriberIntegration(t *testing.T) {
	// !!! Important !!!: run this test with care and only manually.
	if true {
		t.Skip("Skipping TestOrdersSubscriberIntegration due to dangerous test")
		return
	}

	if testing.Short() {
		t.Skip("Skipping TestOrdersSubscriberIntegration")
		return
	}

	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())
	ctx := context.Background()

	t.Logf("Starting watch")
	ch, err := ex.WatchOrdersStatuses(context.Background())
	t.Logf("Got watch result")
	assert.NoError(t, err)

	go func() {
		timer := time.NewTimer(time.Second * 20)
		for {
			select {
			case ev, ok := <-ch:
				if ok {
					t.Logf("MSG: %v", ev)
				}
			case <-timer.C:
				t.Log("Timer!")
				return
			}
		}
	}()

	time.Sleep(2 * time.Second)

	symbol := "LTCUSDT"
	symbolPrice, err := ex.GetPrice(ctx, symbol)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	orderPrice := utils.Mul(symbolPrice, utils.FromString("0.8"))
	quantity := utils.FromString("0.1")
	clientOrderID := uuid.Must(uuid.NewV4()).String()

	t.Logf("Placing order -- symbolprice=%v, orderprice=%v, qty=%v, clientOrderID=%v",
		symbolPrice, orderPrice, quantity, clientOrderID)
	id, err := ex.PlaceBuyOrder(ctx, true, symbol, orderPrice, quantity, clientOrderID)
	defer ex.CancelOrder(ctx, symbol, id)

	t.Logf("Order placed")
	assert.NoError(t, err, "Can't place")
	assert.Equal(t, clientOrderID, id)

	time.Sleep(2 * time.Second)

	err = ex.CancelOrder(ctx, symbol, id)
	assert.NoError(t, err, "Can't cancel")
	if err == nil {
		t.Log("Order canceled")
	}

	time.Sleep(2 * time.Second)
}

// go test ./... -run TestDoubleOrderPlacementIntegration -v -timeout 30000ms
func TestDoubleOrderPlacementIntegration(t *testing.T) {
	// !!! Important !!!: run this test with care and only manually.
	if true {
		t.Skip("Skipping TestDoubleOrderPlacementIntegration due to dangerous test")
		return
	}

	if testing.Short() {
		t.Skip("Skipping TestDoubleOrderPlacementIntegration")
		return
	}

	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())
	ctx := context.Background()

	symbol := "LTCUSDT"
	symbolPrice, err := ex.GetPrice(ctx, symbol)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	orderPrice := utils.Mul(symbolPrice, utils.FromString("0.7"))
	quantity := utils.FromString("0.1")
	clientOrderID := uuid.Must(uuid.NewV4()).String()

	t.Logf("Placing order -- symbolprice=%v, orderprice=%v, qty=%v, clientOrderID=%v",
		symbolPrice, orderPrice, quantity, clientOrderID)
	id1, err := ex.PlaceBuyOrder(ctx, true, symbol, orderPrice, quantity, clientOrderID)
	defer ex.CancelOrder(ctx, symbol, id1)

	t.Logf("Order placed 1")
	assert.NoError(t, err, "Can't place 1")
	assert.Equal(t, clientOrderID, id1)

	time.Sleep(1 * time.Second)

	id2, err := ex.PlaceBuyOrder(ctx, true, symbol, orderPrice, quantity, clientOrderID)
	defer ex.CancelOrder(ctx, symbol, id2)

	t.Logf("Order placed 2")
	assert.NoError(t, err, "Can't place 2")
	assert.Equal(t, clientOrderID, id2)

	err = ex.CancelOrder(ctx, symbol, id1)
	assert.NoError(t, err, "Can't cancel")
	if err == nil {
		t.Log("Order canceled")
	}
}

// go test ./... -run TestDoubleOrderPlacementV2Integration -v -timeout 30000ms
func TestDoubleOrderPlacementV2Integration(t *testing.T) {
	// !!! Important !!!: run this test with care and only manually.
	if true {
		t.Skip("Skipping TestDoubleOrderPlacementV2Integration due to dangerous test")
		return
	}

	if testing.Short() {
		t.Skip("Skipping TestDoubleOrderPlacementV2Integration")
		return
	}

	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())
	ctx := context.Background()

	symbol := "LTCUSDT"
	symbolPrice, err := ex.GetPrice(ctx, symbol)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	orderPrice := utils.Mul(symbolPrice, utils.FromString("0.7"))
	quantity := utils.FromString("0.1")
	clientOrderID := uuid.Must(uuid.NewV4()).String()

	t.Logf("Placing order v2-- symbolprice=%v, orderprice=%v, qty=%v, clientOrderID=%v",
		symbolPrice, orderPrice, quantity, clientOrderID)
	id1, err := ex.PlaceBuyOrderV2(ctx, true, symbol, orderPrice, quantity, clientOrderID, "STOP_LOSS")
	defer ex.CancelOrder(ctx, symbol, id1)

	t.Logf("Order placed 1")
	assert.NoError(t, err, "Can't place 1")
	assert.Equal(t, clientOrderID, id1)

	time.Sleep(1 * time.Second)

	id2, err := ex.PlaceBuyOrderV2(ctx, true, symbol, orderPrice, quantity, clientOrderID, "STOP_LOSS")
	defer ex.CancelOrder(ctx, symbol, id2)

	t.Logf("Order placed 2")
	assert.NoError(t, err, "Can't place 2")
	assert.Equal(t, clientOrderID, id2)

	err = ex.CancelOrder(ctx, symbol, id1)
	assert.NoError(t, err, "Can't cancel")
	if err == nil {
		t.Log("Order canceled")
	}
}

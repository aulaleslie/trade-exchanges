package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strconv"

	api "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Binance struct {
	client *api.Client
	ex     exchanges.Exchange
}

func main() {
	key, _ := os.LookupEnv("KEY")
	secret, _ := os.LookupEnv("SECRET")
	client := api.NewClient(key, secret)

	ex := binance.NewBinanceLong(binance.OriginalBinanceURLs, key, secret, zap.NewExample())
	bn := &Binance{client, ex}

	bn.CancelOrder()
}

func (b *Binance) CancelOrder() {
	symbol := "LTCUSDT"
	clientOrderID := "RUN76-b40bfc43" // CANCELED
	// clientOrderID := "RUN76-12323123123" // NOT FOUND
	// clientOrderID := "RUN166-c213443c" // FILLED
	orderID := int64(1349680305)
	_, _ = clientOrderID, orderID

	err := b.CancelOrderInternal(symbol, clientOrderID)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Print("Canceled")
}

func (b *Binance) CancelOrderInternal(symbol, clientOrderID string) error {
	log.Print("CancelOrderInternal")
	err := b.sendCancelOrderRequest(symbol, clientOrderID)
	switch {
	case err == nil:
		return nil
	case err != exchanges.OrderNotFoundError:
		return err
	}

	status, err := b.GetOrderStatusInternal(symbol, clientOrderID)
	if err != nil {
		return err
	}

	switch status {
	case api.OrderStatusTypeNew, api.OrderStatusTypePartiallyFilled:
		return errors.Errorf("Can't cancel order + order nave status = %v on Binance", status)
	case api.OrderStatusTypeFilled:
		return exchanges.OrderExecutedError
	case api.OrderStatusTypeCanceled:
		return nil
	case api.OrderStatusTypePendingCancel:
		return errors.New("Order have PENDING_CANCEL status.")
	case api.OrderStatusTypeRejected:
		return nil
	case api.OrderStatusTypeExpired:
		return nil
	default:
		return errors.Errorf("Can't cancel order + order nave unknown status = %v on Binance", status)
	}
}

func (b *Binance) sendCancelOrderRequest(symbol, clientOrderID string) error {
	log.Print("sendCancelOrderRequest")
	ctx := context.Background()
	_, err := b.client.NewCancelOrderService().
		Symbol(symbol).
		OrigClientOrderID(clientOrderID).
		Do(ctx)
	if err == nil {
		return nil
	}
	if isNotFoundDuringCancellation(err) {
		return exchanges.OrderNotFoundError
	}

	return err
}

func isNotFoundDuringCancellation(err error) bool {
	if err == nil {
		return false
	}

	// https://binance-docs.github.io/apidocs/spot/en/#order-rejection-issues
	apiErr, ok := err.(*common.APIError)
	if !ok {
		log.Printf("isNotFoundDuringCancellation nonAPIError; T(err) = %s, err = '%v'", reflect.TypeOf(err).String(), err)
		return false
	}

	return apiErr.Code == -2011 && apiErr.Message == "Unknown order sent."
}

func (b *Binance) GetOrderStatusInternal(symbol, clientOrderID string) (api.OrderStatusType, error) {
	log.Printf("GetOrderStatusInternal")
	ctx := context.Background()
	order, err := b.client.NewGetOrderService().
		Symbol(symbol).
		OrigClientOrderID(clientOrderID).
		Do(ctx)
	if err != nil {
		if isNotFoundDuringGetOrderStatus(err) {
			return "", exchanges.OrderNotFoundError
		}
		return "", err
	}

	return order.Status, nil
}

func isNotFoundDuringGetOrderStatus(err error) bool {
	log.Print("isNotFoundDuringGetOrderStatus")
	// https://binance-docs.github.io/apidocs/spot/en/#11xx-2xxx-request-issues
	if err == nil {
		log.Printf("isNotFoundDuringGetOrderStatus nil")
		return false
	}

	apiErr, ok := err.(*common.APIError)
	if !ok {
		log.Printf("isNotFoundDuringGetOrderStatus nonAPIError; T(err) = %s, err = '%v'", reflect.TypeOf(err).String(), err)
		return false
	}

	log.Printf("isNotFoundDuringGetOrderStatus '%d'", apiErr.Code)
	return apiErr.Code == -2013
}

func printAsJSON(x interface{}, err error) {
	fatalIfErr(err)

	text, err := json.MarshalIndent(x, "", " ")
	if err != nil {
		log.Fatalf("JSON Error: %v", err)
	}
	log.Print(string(text))
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func mustParseInt64(s string) int64 {
	x, err := strconv.ParseInt(s, 10, 64)
	fatalIfErr(err)
	return x
}

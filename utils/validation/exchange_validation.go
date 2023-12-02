package validation

import (
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

func ValidateBinancePrice(orderPrice, orderQuantity *apd.Decimal, priceFilter, notion, lotSize map[string]interface{}) error {
	// Validation base on binance us documentation. For details please refer to: https://docs.binance.us/#filters
	// Price
	tickSize := utils.FromString(priceFilter["tickSize"].(string))
	maxPrice := utils.FromString(priceFilter["maxPrice"].(string))
	minNotion := utils.FromString(notion["minNotional"].(string))
	rem := utils.Mod(orderPrice, tickSize)

	// Validate price tick size
	if utils.Greater(rem, apd.New(0, 0)) {
		return errors.Errorf("invalid price price : %v, price tickSize: %v", orderPrice, tickSize)
	}
	// Validate max price
	if utils.Greater(orderPrice, maxPrice) {
		return errors.Errorf("orderPrice exceed the maxPrice allowed. orderPrice: %v, maxPrice: %v", orderPrice, maxPrice)
	}
	// Validate minimum notion
	orderNotion := utils.Mul(orderPrice, orderQuantity)
	if utils.Less(orderNotion, minNotion) {
		return errors.Errorf("invalid request. order notion bot fullfil the minNotion. OrderNotion : %v, minNotion: %v", orderNotion, minNotion)
	}
	return nil
}

func ValidateBinanceQuantity(orderQuantity *apd.Decimal, priceFilter, notion, lotSize map[string]interface{}) error {
	// Validation base on binance us documentation. For details please refer to: https://docs.binance.us/#filters
	// Quantity
	minQty := utils.FromString(lotSize["minQty"].(string))
	maxQty := utils.FromString(lotSize["maxQty"].(string))
	stepSize := utils.FromString(lotSize["stepSize"].(string))

	// Validate quantity step size
	remQtt := utils.Mod(utils.Sub(orderQuantity, minQty), stepSize)
	if utils.Greater(remQtt, apd.New(0, 0)) {
		return errors.Errorf("invalid quantity : %v, quantity stepSize: %v", orderQuantity, stepSize)
	}
	// Validate maximum quantity size
	if utils.Greater(orderQuantity, maxQty) {
		return errors.Errorf("quantity exceed the maxQty allowed quantity: %v, maxQty: %v", orderQuantity, maxQty)
	}
	return nil
}

func ValidatePhemexPrice(orderPrice *apd.Decimal, lotSizeString, tickSizeString string) error {
	tickSize := utils.FromString(tickSizeString)
	//Price vallidation
	remPrice := utils.Mod(orderPrice, tickSize)
	if utils.Greater(remPrice, apd.New(0, 0)) {
		return errors.Errorf("invalid orderPrice : %v, orderPrice tickSize: %v", orderPrice, tickSize)
	}
	return nil
}

func ValidatePhemexQuantity(orderQuantity *apd.Decimal, lotSizeString, tickSizeString string) error {
	lotSize := utils.FromString(lotSizeString)
	// Quantity
	// Validate min quantity size
	if utils.Less(orderQuantity, lotSize) {
		return errors.Errorf("quantity size is not allowed. quantity: %v, min quantity: %v", orderQuantity, lotSize)
	}
	return nil
}

func RoundBinancePriceAndQuantity(orderPrice, orderQuantity *apd.Decimal, priceFilter, lotSize map[string]interface{}) (
	price, qtt *apd.Decimal, err error) {
	// Validation base on binance us documentation. For details please refer to: https://docs.binance.us/#filters
	// Price
	price = orderPrice
	qtt = orderQuantity

	tickSize := utils.FromString(priceFilter["tickSize"].(string))
	priceRem := utils.Mod(orderPrice, tickSize)
	if tickSize.Exponent < 0 && orderPrice.Exponent < 0 && utils.Greater(priceRem, apd.New(0, 0)) {
		precision := uint32(-tickSize.Exponent)
		priceFloat, err := orderPrice.Float64()
		if err != nil {
			return nil, nil, err
		}
		p := utils.RoundFloat(priceFloat, precision)
		price = utils.SetFloat64(p)
	}
	// Quantity
	stepSize := utils.FromString(lotSize["stepSize"].(string))
	quantityRem := utils.Mod(orderQuantity, stepSize)
	if stepSize.Exponent < 0 && orderQuantity.Exponent < 0 && utils.Greater(quantityRem, apd.New(0, 0)) {
		precision := uint32(-stepSize.Exponent)
		qttFloat, err := orderQuantity.Float64()
		if err != nil {
			return nil, nil, err
		}
		q := utils.FloorFloat(qttFloat, precision)
		qtt = utils.SetFloat64(q)
	}

	return price, qtt, err
}

func RoundPhemexPriceAndQuantity(orderPrice, orderQuantity *apd.Decimal, lotSizeString, tickSizeString string) (
	price, qtt *apd.Decimal, err error) {
	lotSize := utils.FromString(lotSizeString)
	tickSize := utils.FromString(tickSizeString)
	price = orderPrice
	qtt = orderQuantity
	//Price vallidation
	remPrice := utils.Mod(orderPrice, tickSize)
	if tickSize.Exponent < 0 && orderPrice.Exponent < 0 && utils.Greater(remPrice, apd.New(0, 0)) {
		precision := uint32(-tickSize.Exponent)
		priceFloat, err := orderPrice.Float64()
		if err != nil {
			return nil, nil, err
		}
		p := utils.RoundFloat(priceFloat, precision)
		price = utils.SetFloat64(p)
	}
	// Quantity
	remQtt := utils.Mod(orderQuantity, lotSize)
	if lotSize.Exponent < 0 && orderQuantity.Exponent < 0 && utils.Greater(remQtt, apd.New(0, 0)) {
		precision := uint32(-lotSize.Exponent)
		qttFloat, err := orderQuantity.Float64()
		if err != nil {
			return nil, nil, err
		}
		q := utils.FloorFloat(qttFloat, precision)
		qtt = utils.SetFloat64(q)
	}
	return price, qtt, err
}

func RoundBybitPriceAndQuantity(orderPrice, orderQuantity *apd.Decimal, lotSizeString, tickSizeString string) (
	price, qtt *apd.Decimal, err error) {
	lotSize := utils.FromString(lotSizeString)
	tickSize := utils.FromString(tickSizeString)
	price = orderPrice
	qtt = orderQuantity
	//Price vallidation
	remPrice := utils.Mod(orderPrice, tickSize)
	if tickSize.Exponent < 0 && orderPrice.Exponent < 0 && utils.Greater(remPrice, apd.New(0, 0)) {
		precision := uint32(-tickSize.Exponent)
		priceFloat, err := orderPrice.Float64()
		if err != nil {
			return nil, nil, err
		}
		p := utils.RoundFloat(priceFloat, precision)
		price = utils.SetFloat64(p)
	}
	// Quantity
	remQtt := utils.Mod(orderQuantity, lotSize)
	if lotSize.Exponent < 0 && orderQuantity.Exponent < 0 && utils.Greater(remQtt, apd.New(0, 0)) {
		precision := uint32(-lotSize.Exponent)
		qttFloat, err := orderQuantity.Float64()
		if err != nil {
			return nil, nil, err
		}
		q := utils.FloorFloat(qttFloat, precision)
		qtt = utils.SetFloat64(q)
	}
	return price, qtt, err
}

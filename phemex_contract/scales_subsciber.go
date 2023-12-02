package phemex_contract

import (
	"context"
	"time"

	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var ScalesSubscriberInstance *ScalesSubscriber

func InitScalesSubscriber(lg *zap.Logger) {
	ScalesSubscriberInstance = NewScalesSubscriber(lg)
	ScalesSubscriberInstance.Start(context.Background())
}

type SymbolScale struct {
	PriceScale        int
	PriceScaleDivider *apd.Decimal
}

// TODO: it's better to store scales in persistent storage
type ScalesSubscriber struct {
	scales atomic.Value // map[string]SymbolScale symbol->scale
	lg     *zap.Logger
}

func NewScalesSubscriber(lg *zap.Logger) *ScalesSubscriber {
	return &ScalesSubscriber{
		lg: lg.Named("ScalesSubscriber"),
	}
}

func (ss *ScalesSubscriber) GetLastScales() map[string]SymbolScale {
	last := ss.scales.Load()
	if last == nil {
		return nil
	}
	return last.(map[string]SymbolScale)
}

func (ss *ScalesSubscriber) GetLastSymbolScales(symbol string) (SymbolScale, error) {
	scales := ss.GetLastScales()
	if scales == nil {
		return SymbolScale{}, errors.Errorf("scales are empty")
	}

	symbolScales, ok := scales[symbol]
	if !ok {
		return SymbolScale{}, errors.Errorf("scales for symbol %s not found", symbol)
	}
	return symbolScales, nil
}

func (ss *ScalesSubscriber) storeLastScales(last map[string]SymbolScale) {
	ss.scales.Store(last)
}

func (ss *ScalesSubscriber) fetchScales(ctx context.Context) (map[string]SymbolScale, error) {
	// Rate limiter headers are present here but b/c the call are rare and the singleton pattern it's ok to have no ratelimiter.
	data, err := krisa_phemex_fork.NewClient("", "", ss.lg).NewProductsService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to make request")
	}

	result := map[string]SymbolScale{}
	for _, product := range data.Products {
		if product.Type != krisa_phemex_fork.PerpetualProductType {
			continue
		}

		priceScaleDivider, err := ScaleToDivider(int(product.PriceScale))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid price scale for %s", product.Symbol)
		}
		result[product.Symbol] = SymbolScale{
			PriceScale:        int(product.PriceScale),
			PriceScaleDivider: priceScaleDivider,
		}
	}
	return result, nil
}

func (ss *ScalesSubscriber) CheckOrUpdate(ctx context.Context) error {
	if ss.GetLastScales() != nil {
		return nil
	}

	scales, err := ss.fetchScales(ctx)
	if err != nil {
		return err
	}

	ss.lg.Info("Updating Phemex scales")
	ss.storeLastScales(scales)
	return nil
}

func (ss *ScalesSubscriber) Start(ctx context.Context) {
	go func() {
		ss.lg.Info("Starting Phemex scales subscriber")

		sequentialErrors := 0
		sequentialErrorsLimit := 5
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		fn := func() {
			scales, err := ss.fetchScales(ctx)
			if err != nil {
				sequentialErrors++
				ss.lg.Error("Can't fetch Phemex scales",
					zap.Int("sequentialErrors", sequentialErrors), zap.Error(err))
				if sequentialErrors > sequentialErrorsLimit {
					ss.lg.Info("Erasing Phemex scales")
					ss.storeLastScales(nil)
				}
			} else {
				sequentialErrors = 0
				ss.lg.Info("Sucessful fetching Phemex scales")
				ss.storeLastScales(scales)
			}
		}

		fn()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fn()
			}
		}
	}()
}

//////////////////////////

func ScaleToDivider(scale int) (*apd.Decimal, error) {
	if scale < 0 {
		return nil, errors.New("negative scale")
	}
	var result *apd.Decimal = utils.FromUint(1)
	for i := 0; i < scale; i++ {
		result = utils.Mul(result, utils.D10)
	}
	return result, nil
}

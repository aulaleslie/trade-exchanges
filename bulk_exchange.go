package exchanges

import (
	"context"

	"github.com/pkg/errors"
)

type SequentialBulkCancelExchange struct {
	Exchange
}

// var _ BulkCancelExchange = (*SequentialBulkCancelExchange)(nil)

func (this *SequentialBulkCancelExchange) BulkCancelOrder(
	ctx context.Context, symbol string, ids []string,
) ([]BulkCancelResult, error) {
	result := []BulkCancelResult{}
	// OPTIMIZATION: make parallel
	for _, id := range ids {
		err := this.CancelOrder(ctx, symbol, id)
		err = errors.Wrapf(err, "can't cancel order (OrderID=%s)", id)
		result = append(result, BulkCancelResult{ID: id, Err: err})
	}
	return result, nil
}

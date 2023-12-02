package futures

import (
	"context"
	"fmt"

	api "github.com/adshao/go-binance/v2/futures"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/cockroachdb/apd"
)

type PositionGetter struct {
	client *api.Client
}

func NewPositionGetter(client *api.Client) *PositionGetter {
	return &PositionGetter{
		client: client,
	}
}

func (pg *PositionGetter) getBinanceAccount(ctx context.Context) (*api.Account, error) {
	res, err := pg.client.NewGetAccountService().Do(ctx)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res, nil
}

func (pg *PositionGetter) GetAccountPosition(ctx context.Context) (res exchanges.Account, err error) {
	account, err := pg.getBinanceAccount(ctx)
	if err != nil {
		return
	}

	accountPositions := make([]exchanges.AccountPosition, 0)
	for _, position := range account.Positions {
		unrealizedProfit, _, err := apd.NewFromString(position.UnrealizedProfit)
		if err != nil {
			return res, err
		}

		leverage, _, err := apd.NewFromString(position.Leverage)
		if err != nil {
			return res, err
		}

		accountPosition := exchanges.AccountPosition{
			Symbol:           position.Symbol,
			UnrealizedProfit: unrealizedProfit,
			Leverage:         leverage,
			EntryPrice:       nil,
			Size:             nil,
		}

		accountPositions = append(accountPositions, accountPosition)
	}

	res.AccountPositions = accountPositions

	return
}

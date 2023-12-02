package binance

import (
	"context"
	"fmt"

	api "github.com/adshao/go-binance/v2"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/cockroachdb/apd"
)

type PositionGetter struct {
	client *api.Client
}

func (pg *PositionGetter) getBinanceAccount(ctx context.Context) (*api.Account, error) {
	res, err := pg.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res, nil
}

func (pg *PositionGetter) GetAccountBalances(ctx context.Context) (res exchanges.Account, err error) {
	account, err := pg.getBinanceAccount(ctx)
	if err != nil {
		return
	}

	accountBalances := make([]exchanges.AccountBalance, 0)
	for _, balance := range account.Balances {
		free, _, err := apd.NewFromString(balance.Free)
		if err != nil {
			return res, err
		}

		locked, _, err := apd.NewFromString(balance.Locked)
		if err != nil {
			return res, err
		}

		accountBalance := exchanges.AccountBalance{
			Coin:   balance.Asset,
			Free:   free,
			Locked: locked,
		}

		accountBalances = append(accountBalances, accountBalance)
	}

	res.AccountBalances = accountBalances

	return
}

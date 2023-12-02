package phemex_contract

import (
	"context"
	"fmt"

	"github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
)

type PositionFetcher struct {
	client *phemex.Client
}

func NewPositionFetcher(client *phemex.Client) *PositionFetcher {
	return &PositionFetcher{
		client: client,
	}
}

func (p *PositionFetcher) GetAccountPosition(ctx context.Context) (res exchanges.Account, err error) {
	accountPosition, err := p.client.NewGetAccountPositionService().Do(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	if accountPosition == nil {
		return
	}

	accountPositions := make([]exchanges.AccountPosition, 0)
	for _, position := range accountPosition.Positions {
		unrealizedProfit := utils.FromFloat64(position.UnRealisedPosLoss)
		leverage := utils.FromFloat64(position.Leverage)
		entryPrice := utils.FromFloat64(position.AvgEntryPrice)
		size := utils.FromFloat64(position.Size)

		accountPosition := exchanges.AccountPosition{
			Symbol:           position.Symbol,
			UnrealizedProfit: unrealizedProfit,
			Leverage:         leverage,
			EntryPrice:       entryPrice,
			Size:             size,
		}

		accountPositions = append(accountPositions, accountPosition)
	}
	res.AccountPositions = accountPositions

	accountBalances := make([]exchanges.AccountBalance, 0)
	res.AccountBalances = accountBalances

	return res, nil
}

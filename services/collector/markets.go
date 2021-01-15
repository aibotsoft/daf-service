package collector

import (
	"context"
	"github.com/aibotsoft/daf-service/pkg/store"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

func (c *Collector) CollectMarketsJob() {
	for {
		time.Sleep(time.Hour * 24)
		err := c.CollectMarketsRound()
		if err != nil {
			c.log.Info(err)
		}
	}
}

func (c *Collector) CollectMarketsRound() error {
	ctx := context.Background()
	resp, _, err := c.client.OddsApi.GetBetTypes(ctx).Lang("en-US").Execute()
	if err != nil {
		return err
	}
	var markets []store.Market
	for betTypeIdStr, betTypeName := range resp {
		betTypeId, err := strconv.ParseInt(betTypeIdStr, 10, 64)
		if err != nil {
			c.log.Error(err)
			continue
		}
		markets = append(markets, store.Market{Id: betTypeId, Name: betTypeName})
	}
	err = c.store.SaveMarkets(ctx, markets)
	if err != nil {
		return errors.Wrap(err, "save markets error")
	}
	return nil
}

package handler

import (
	"context"
	"database/sql"
	"github.com/aibotsoft/daf-service/pkg/api"
	"github.com/aibotsoft/gen/fortedpb"
)

func (h *Handler) processLine(ctx context.Context, side *fortedpb.SurebetSide, line *api.Ticket) error {
	if side.MarketId == 0 {
		line.EventId = side.EventId
		var err error
		err = h.store.FindLine(ctx, line)

		h.log.Info(line)
		switch err {
		case nil:
			side.MarketId = line.Id
			return nil
		case sql.ErrNoRows:
			for _, state := range eventStates {
				event := &api.Event{Id: side.EventId, Home: side.Home, Away: side.Away, LeagueId: side.LeagueId, SportId: side.SportId, EventState: state}
				got, err := h.api.GetMarkets(ctx, h.auth.Base(), event)
				if err != nil {
					return err
				}
				err = h.store.SaveLines(ctx, got)
				if err != nil {
					return err
				}
			}
		default:
			return err
		}

		err = h.store.FindLine(ctx, line)
		if err != nil {
			return err
		}
		side.MarketId = line.Id
	}
	line.Id = side.MarketId
	return nil
}

package handler

import (
	"context"
	"database/sql"
	"github.com/aibotsoft/gen/fortedpb"
)

func (h *Handler) processEvent(ctx context.Context, side *fortedpb.SurebetSide) error {
	if side.EventId == 0 {
		var err error
		for i := 0; i < 2; i++ {
			side.EventId, err = h.store.FindEventByName(ctx, side.Home, side.Away, side.LeagueId, side.SportId)
			switch err {
			case nil:
				return nil
			case sql.ErrNoRows:
				for _, state := range eventStates {
					got, err := h.api.GetEvents(ctx, h.auth.Base(), state, side.SportId, side.LeagueId)
					if err != nil {
						return err
					}
					err = h.store.SaveEvents(ctx, got)
					if err != nil {
						return err
					}
				}
			default:
				return err
			}
		}
	}
	return nil
}

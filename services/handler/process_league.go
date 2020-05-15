package handler

import (
	"context"
	"database/sql"
	"github.com/aibotsoft/gen/fortedpb"
)

func (h *Handler) processLeague(ctx context.Context, side *fortedpb.SurebetSide) error {
	if side.LeagueId == 0 {
		var err error
		side.LeagueId, err = h.store.FindLeagueByName(ctx, side.SportId, side.LeagueName)
		switch err {
		case nil:
			return nil
		case sql.ErrNoRows:
			for _, state := range eventStates {
				got, err := h.api.GetLeagues(ctx, h.auth.Base(), state, side.SportId)
				if err != nil {
					return err
				}
				err = h.store.SaveLeagues(ctx, got)
				if err != nil {
					return err
				}
			}
		default:
			return err
		}
		side.LeagueId, err = h.store.FindLeagueByName(ctx, side.SportId, side.LeagueName)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

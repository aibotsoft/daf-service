package handler

import (
	"context"
	"database/sql"
	"github.com/aibotsoft/gen/fortedpb"
)

func (h *Handler) processSport(ctx context.Context, side *fortedpb.SurebetSide) error {
	if side.SportId == 0 {
		var err error
		side.SportId, err = h.store.FindSportByName(ctx, side.SportName)
		switch err {
		case nil:
			return nil
		case sql.ErrNoRows:
			for _, state := range eventStates {
				got, err := h.api.GetSports(ctx, h.auth.Base(), state)
				if err != nil {
					return err
				}
				err = h.store.SaveSports(ctx, got)
				if err != nil {
					return err
				}
			}
		default:
			return err
		}
		side.SportId, err = h.store.FindSportByName(ctx, side.SportName)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

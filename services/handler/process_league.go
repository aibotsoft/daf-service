package handler

import (
	"context"
	"github.com/aibotsoft/gen/fortedpb"
)

func (h *Handler) processLeague(ctx context.Context, side *fortedpb.SurebetSide) (err error) {
	if side.LeagueId == 0 {
		side.LeagueId, err = h.store.FindLeagueByName(ctx, side.SportId, side.LeagueName)
	}
	return
}

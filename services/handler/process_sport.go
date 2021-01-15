package handler

import (
	"context"
	"github.com/aibotsoft/gen/fortedpb"
)

func (h *Handler) processSport(ctx context.Context, side *fortedpb.SurebetSide) (err error) {
	if side.SportId == 0 {
		side.SportId, err = h.store.FindSportByName(ctx, side.SportName)
	}
	return
}

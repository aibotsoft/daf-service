package handler

import (
	pb "github.com/aibotsoft/gen/fortedpb"
)

//func (h *Handler) GetCurrency(sb *fortedpb.Surebet) float64 {
//	for _, currency := range sb.Currency {
//		if currency.Code == h.auth.Account.CurrencyCode {
//			return currency.Value
//		}
//	}
//	return 0
//}

func (h *Handler) GetCurrency(sb *pb.Surebet) float64 {
	for _, currency := range sb.Currency {
		if currency.Code == h.auth.Account.CurrencyCode {
			return currency.Value
		}
	}
	return 0
}

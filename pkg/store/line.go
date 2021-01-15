package store

import (
	"context"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
	"sync"
)

type Ticket struct {
	Id      int64
	BetTeam string
	Price   float64

	BetTypeId int64
	Points    *float64
	EventId   int64

	MarketName string
	IsLive     bool
	MinBet     float64
	MaxBet     float64
	Home       string
	Away       string
	Cat        int64
	IsHandicap bool
	EventState string
}

func (t *Ticket) CompareCat(that int64) bool {
	if t.Cat == 0 {
		return true
	}
	return t.Cat == that
}
func (t *Ticket) ComparePoints(that *float64) bool {
	if t.Points == nil {
		return true
	}
	if that == nil {
		return false
	}
	return *t.GetDbPoints() == *that
}

func (t *Ticket) GetPoints() float64 {
	if t.Points == nil {
		return 0
	}
	return *t.Points
}

func (t *Ticket) GetSidePoints() float64 {
	if t.Points == nil {
		return 0
	}
	return *t.Points
}

//В базе данных гандикап хранится для второй команды, поэтому нужно возвращать его как есть для второй,
//и инвертировать знак для первой команды (только форы)
func (t *Ticket) GetDbPoints() *float64 {
	if t.Points == nil {
		return nil
	}
	if !t.IsHandicap {
		return t.Points
	}
	if t.BetTeam == "a" {
		return t.Points
	}
	p := *t.Points * -1
	return &p
}

func (t *Ticket) Equal(that interface{}) bool {
	if that == nil {
		return t == nil
	}
	that1, ok := that.(*Ticket)
	if !ok {
		that2, ok := that.(Ticket)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return t == nil
	} else if t == nil {
		return false
	}
	if t.BetTypeId != that1.BetTypeId {
		return false
	}
	if t.Points == nil {
		return that1.Points == nil
	}
	if that1.Points == nil {
		return t.Points == nil
	}
	if *t.Points != *that1.Points {
		return false
	}
	if t.EventId != that1.EventId {
		return false
	}
	//if t.Cat != that1.Cat {
	//	return false
	//}
	return true
}

var ticketCache sync.Map

func (s *Store) SaveLines(ctx context.Context, tickets []Ticket) error {
	var newTickets []Ticket
	for i := range tickets {
		//if tickets[i].Id == 251461446 {
		//	s.log.Infow("", "", tickets[i])
		//}
		got, ok := ticketCache.Load(tickets[i].Id)
		if !ok {
			newTickets = append(newTickets, tickets[i])
			ticketCache.Store(tickets[i].Id, tickets[i])
		} else if !tickets[i].Equal(got) {
			newTickets = append(newTickets, tickets[i])
			ticketCache.Store(tickets[i].Id, tickets[i])
			//s.log.Infow("ticket not equal", "first", got, "second", tickets[i])
		}
	}
	if len(newTickets) == 0 {
		//s.log.Debug("no newTickets")
		return nil
	}
	s.log.Debugw("got new lines", "count", len(newTickets))
	tvp := mssql.TVP{TypeName: "LineType", Value: newTickets}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateLines", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateLines error ")
	}
	return nil
}

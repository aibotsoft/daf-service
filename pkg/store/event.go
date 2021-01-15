package store

import (
	"context"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
	"sync"
)

type Event struct {
	Id         int64
	Home       int64
	Away       int64
	LeagueId   int64
	SportId    int64
	EventState string
	Starts     string
}

func (t *Event) Equal(that interface{}) bool {
	if that == nil {
		return t == nil
	}
	that1, ok := that.(*Event)
	if !ok {
		that2, ok := that.(Event)
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
	if t.Home != that1.Home {
		return false
	}
	if t.Away != that1.Away {
		return false
	}
	if t.LeagueId != that1.LeagueId {
		return false
	}
	if t.SportId != that1.SportId {
		return false
	}
	if t.Starts != that1.Starts {
		return false
	}
	return true
}

var eventCache sync.Map

func (s *Store) SaveEvents(ctx context.Context, events []Event) error {
	var newEvents []Event
	for i := range events {
		//if events[i].Home == "Ilya Novikov" {
		//
		//}
		got, ok := eventCache.Load(events[i].Id)
		if !ok {
			newEvents = append(newEvents, events[i])
			eventCache.Store(events[i].Id, events[i])
		} else if !events[i].Equal(got) {
			newEvents = append(newEvents, events[i])
			eventCache.Store(events[i].Id, events[i])
		}
	}
	if len(newEvents) == 0 {
		//s.log.Debug("no newEvents")
		return nil
	}
	s.log.Debugw("got new events", "count", len(newEvents))

	tvp := mssql.TVP{TypeName: "EventType", Value: newEvents}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateEvents", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateEvents error for %v", newEvents)
	}
	return nil
}

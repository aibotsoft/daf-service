package store

import (
	"context"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/pkg/errors"
	"sync"
)

type Sport struct {
	Id         int64
	Name       string
	Count      int
	EventState string
}

func (t *Sport) Equal(that interface{}) bool {
	if that == nil {
		return t == nil
	}
	that1, ok := that.(*Sport)
	if !ok {
		that2, ok := that.(Sport)
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
	if t.Name != that1.Name {
		return false
	}
	return true
}

var sportCache sync.Map

func (s *Store) SaveSports(ctx context.Context, sports []Sport) error {
	var newSports []Sport
	for i := range sports {
		got, ok := sportCache.Load(sports[i].Id)
		if !ok {
			newSports = append(newSports, sports[i])
			sportCache.Store(sports[i].Id, sports[i])
		} else if !sports[i].Equal(got) {
			newSports = append(newSports, sports[i])
			sportCache.Store(sports[i].Id, sports[i])
		}
	}
	if len(newSports) == 0 {
		//s.log.Debug("no newSports")
		return nil
	}
	s.log.Debugw("got new sports", "count", len(newSports))

	tvp := mssql.TVP{TypeName: "SportType", Value: newSports}
	_, err := s.db.ExecContext(ctx, "dbo.uspCreateSports", tvp)
	if err != nil {
		return errors.Wrapf(err, "uspCreateSport error for %v", newSports)
	}
	return nil
}

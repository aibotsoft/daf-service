package api

import (
	"context"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Event struct {
	Id         int64
	Home       string
	Away       string
	LeagueId   int64
	SportId    int64
	EventState string
}

var eventsPath, _ = url.Parse(`match.aspx`)
var homeAwayRe = regexp.MustCompile(`(.*)\s+-vs-\s+(.*)`)

func HomeAwaySplit(eventStr string) (string, string, error) {
	match := homeAwayRe.FindStringSubmatch(eventStr)
	if len(match) < 3 {
		return "", "", errors.Errorf("not found teams in event string %v", eventStr)
	}
	return strings.TrimSpace(match[1]), strings.TrimSpace(match[2]), nil
}
func FindEventId(node *html.Node) (int64, error) {
	u, err := url.Parse(htmlquery.SelectAttr(node, "href"))
	if err != nil {
		return 0, err
	}
	q := u.Query()
	matchId := q["matchid"]
	if len(matchId) == 0 {
		return 0, errors.Errorf("not found eventId")
	}
	id, err := strconv.ParseInt(matchId[0], 10, 64)
	if err != nil {
		return 0, errors.Errorf("not found eventId")
	}
	return id, nil
}
func (a *Api) GetEvents(ctx context.Context, base *url.URL, eventState string, sportId int64, leagueId int64) ([]Event, error) {
	var events []Event
	resp, err := a.client.R().EnableTrace().SetContext(ctx).SetDoNotParseResponse(true).
		SetQueryParam("eventState", eventState).
		SetQueryParam("sport", strconv.FormatInt(sportId, 10)).
		SetQueryParam("leagueID", strconv.FormatInt(leagueId, 10)).
		Get(base.ResolveReference(eventsPath).String())
	if err != nil {
		return nil, errors.Wrapf(err, "get events error for sport %v and league %v", sportId, leagueId)
	}
	page, err := parse(resp)
	if err != nil {
		return nil, err
	}
	trace := resp.Request.TraceInfo()
	if !trace.IsConnReused {
		a.log.Infow("conn not reused", "trace", trace)
	}
	links, err := htmlquery.QueryAll(page, `//a[contains(@href,'underover/oddsnew.aspx')]`)
	if err != nil {
		return nil, errors.Wrapf(err, "query //a in event page error for sport %v and league %v", sportId, leagueId)
	}
	for _, link := range links {
		eventId, err := FindEventId(link)
		if err != nil {
			a.log.Warn(err)
			continue
		}

		home, away, err := HomeAwaySplit(htmlquery.InnerText(link))
		if err != nil {
			a.log.Warn(err)
			continue
		}
		events = append(events, Event{
			Id:         eventId,
			Home:       home,
			Away:       away,
			SportId:    sportId,
			LeagueId:   leagueId,
			EventState: eventState,
		})
	}
	return events, nil
}

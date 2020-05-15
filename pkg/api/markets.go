package api

import (
	"context"
	"github.com/pkg/errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const NoOdds = "No odds data"

var eventPath, _ = url.Parse(`underover/oddsnew.aspx`)

//var brSplitRe = regexp.MustCompile(`<br>\s*<br>`)
//var linkRe = regexp.MustCompile(`ticket\.aspx\?id=(\d+)&odds=(.*)&option=(\w)&isor=(\w)&islive=(\w)">(.*)@(.*)<\/a>`)
var pointRe = regexp.MustCompile(`.*\s+(-?\d+\.?\d*)\s*`)
var marketSplitRe = regexp.MustCompile(`<b>\[(.*)\]</b><br>\r?\n(<a.*\r?\n)+`)
var urlRe = regexp.MustCompile(`BetNew.aspx?(.*?)">`)
var priceRe = regexp.MustCompile(`">\s?(.*?)\s?@\s?(.*?)\s?</a>`)

//var teamsRe = regexp.MustCompile(`<b>\s?(.*?)\s*-vs-\s*(.*?)\s*<br>`)

type Ticket struct {
	Id        int64
	BetTeam   string
	Price     float64
	BetTypeId int64
	Points    *float64
	EventId   int64

	MarketName string
	IsLive     bool
	MinBet     float64
	MaxBet     float64
	Home       string
	Away       string
}

func (a *Api) processMarketResponse(resp string, event *Event) ([]Ticket, error) {
	var tickets []Ticket
	home, away, err := parseTeams(resp)
	if err != nil {
		if strings.Index(resp, NoOdds) != -1 {
			return tickets, nil
		}
		return nil, err
	}
	if event.Home != home {
		return nil, errors.Errorf("home teams not equal, %q != %q ", event.Home, home)
	}
	if event.Away != away {
		return nil, errors.Errorf("away teams not equal, %q != %q ", event.Away, away)
	}
	isLive := strings.Index(resp, "Score") != -1
	//a.log.Infow("teams", "home", home, "away", away)
	markets := marketSplitRe.FindAllStringSubmatch(resp, -1)
	for _, s := range markets {

		marketName := s[1]
		split := strings.Split(s[0], "<br>")
		for _, link := range split[1:] {

			findBetNewUrl := urlRe.FindStringSubmatch(link)
			if len(findBetNewUrl) < 2 {
				continue
			}
			u, err := url.Parse(findBetNewUrl[1])
			if err != nil {
				a.log.Error(err)
				continue
			}
			queryMap := u.Query()

			betType := queryMap["BetType"]
			if len(betType) != 1 {
				a.log.Errorf("BetType error for %v", queryMap["BetType"])
				continue
			}
			betTypeId, err := strconv.ParseInt(betType[0], 10, 64)
			if err != nil {
				a.log.Error(err)
				continue
			}

			oddsid := queryMap["oddsid"]
			if len(oddsid) != 1 {
				a.log.Error("oddsid error")
				continue
			}
			id, err := strconv.ParseInt(oddsid[0], 10, 64)
			if err != nil {
				a.log.Error(err)
				continue
			}
			betTeam := queryMap["betteam"]
			if len(betTeam) != 1 {
				a.log.Error("betTeam error")
				continue
			}
			namePriceFound := priceRe.FindStringSubmatch(link)
			if len(namePriceFound) != 3 {
				a.log.Error("not found name and price in url")
				continue
			}
			price, err := strconv.ParseFloat(namePriceFound[2], 64)
			if err != nil {
				a.log.Error(err)
				continue
			}
			var points *float64
			marketTrimmed := strings.TrimSpace(namePriceFound[1])
			if home != marketTrimmed && away != marketTrimmed {
				//a.log.Infow("team_market", "home", home, "market", namePriceFound[1])

				foundPoints := pointRe.FindStringSubmatch(namePriceFound[1])
				if len(foundPoints) > 1 {
					float, err := strconv.ParseFloat(foundPoints[1], 64)
					if err != nil {
						a.log.Errorw("error parse points", "str", namePriceFound[1], "err", err)
					}
					points = &float
				}
			}
			//if points != nil && *points == 1995 {
			//	a.log.Info(home)
			//}
			t := Ticket{
				Id:         id,
				Price:      price,
				MarketName: marketName,
				BetTypeId:  betTypeId,
				Points:     points,
				BetTeam:    betTeam[0],
				EventId:    event.Id,
				Home:       home,
				Away:       away,
				IsLive:     isLive,
			}
			tickets = append(tickets, t)
			//a.log.Infow("", "", t)
		}
	}
	return tickets, nil
}
func (a *Api) GetMarkets(ctx context.Context, base *url.URL, event *Event) ([]Ticket, error) {
	resp, err := a.client.R().EnableTrace().SetContext(ctx).
		SetQueryParam("eventState", event.EventState).
		SetQueryParam("sport", strconv.FormatInt(event.SportId, 10)).
		SetQueryParam("leagueID", strconv.FormatInt(event.LeagueId, 10)).
		SetQueryParam("matchid", strconv.FormatInt(event.Id, 10)).
		Get(base.ResolveReference(eventPath).String())
	if err != nil {
		return nil, errors.Wrapf(err, "get markets error for event %v", event)
	}
	//noOdds:=strings.Index(resp.String(), "No odds available.")
	tickets, err := a.processMarketResponse(resp.String(), event)
	if err != nil {
		return nil, err
	}
	return tickets, nil
}

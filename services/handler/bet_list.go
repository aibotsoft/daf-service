package handler

import (
	"context"
	"encoding/json"
	"github.com/aibotsoft/daf-service/pkg/store"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/util"
	"github.com/antchfx/htmlquery"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var BetDetailRe = regexp.MustCompile(`ShowPopupResult\(\[(.*?)\]\)`)

const betListJobPeriod = time.Minute * 7
const betResultsMaxDays = 5

type Detail struct {
	TransID     string
	SportID     int64
	LeagueID    int64
	MatchID     int64
	Bettype     string
	WinLoseDate string
	HDP         string
}

func (h *Handler) GetResults(ctx context.Context) ([]pb.BetResult, error) {
	return h.store.GetResults(ctx)
}

func (h *Handler) BetListJob() {
	time.Sleep(10 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		err := h.BetList(ctx)
		cancel()
		if err != nil {
			h.log.Info(err)
		}
		time.Sleep(betListJobPeriod)
	}
}

func (h *Handler) BetList(ctx context.Context) error {
	ctxAuth, err := h.auth.Auth(ctx)
	if err != nil {
		return BadBettingStatusError
	}
	var betList []store.Bet
	for i := -betResultsMaxDays; i < 1; i++ {
		resp, _, err := h.client.BetListApi.GetSettledBetList(ctxAuth).Datatype(8).
			Fdate(time.Now().AddDate(0, 0, i).Format("01/02/2006")).Execute()
		if err != nil {
			h.log.Info("get_bet_list_error: ", err)
			continue
		}
		//h.log.Info(resp)
		bets, err := h.processBets(resp)
		if err != nil {
			h.log.Error(err)
			continue
		}
		betList = append(betList, bets...)
	}
	err = h.store.SaveBetList(ctx, betList)
	if err != nil {
		return err
	}
	return nil
}
func (h *Handler) processBets(resp string) ([]store.Bet, error) {
	page, err := parse(resp)
	if err != nil {
		return nil, err
	}
	betNodes, err := htmlquery.QueryAll(page, `//div[@style="display: block;"]`)
	if err != nil || betNodes == nil {
		return nil, err
	}
	var bets []store.Bet
	for i := range betNodes {
		bet, err := processBetNode(betNodes[i])
		if err != nil {
			h.log.Error(err)
			continue
		}
		//h.log.Infow("", "", bet)
		bets = append(bets, *bet)
	}
	return bets, nil
}

func processBetNode(betNode *html.Node) (*store.Bet, error) {
	//109093368398
	var bet store.Bet
	headerNode, err := htmlquery.Query(betNode, `//div[@class="bets_header"]`)
	footerNode, err := htmlquery.Query(betNode, `//div[@class="bets_footer"]`)
	contentNode, err := htmlquery.Query(betNode, `//div[@class="bets_content touch"]`)
	detail, err := htmlquery.Query(footerNode, `//div[@onclick]`)
	if detail != nil {
		onclick := htmlquery.SelectAttr(detail, "onclick")
		if onclick == "" {
			return nil, errors.Errorf("onclick_empty_error")
		}
		err = parseBetDetail(onclick, &bet)
		if err != nil {
			return nil, err
		}
		//return nil, errors.Errorf("detail_empty_error for footer: %v", htmlquery.OutputHTML(betNode, true))
	} else {
		betIdNode, _ := htmlquery.Query(footerNode, `//div[@class="text-lowlight"][4]`)
		if betIdNode == nil {
			return nil, errors.Errorf("bet_id_error for bet: %v", htmlquery.OutputHTML(betNode, true))
		}
		betIdStr := strings.TrimSpace(strings.Trim(htmlquery.InnerText(betIdNode), "ID:"))
		bet.Id, err = strconv.ParseInt(betIdStr, 10, 64)
		if err != nil {
			return nil, errors.Errorf("parse_bet_id_error for betIdStr: %v", betIdStr)
		}
		//fmt.Println(htmlquery.OutputHTML(betIdNode, true))
	}

	oddsNode, err := htmlquery.Query(headerNode, `//span[@class="text-odds"]`)
	if oddsNode != nil {
		oddsSplit := strings.Split(htmlquery.InnerText(oddsNode), "@")
		if len(oddsSplit) == 2 {
			points, err := strconv.ParseFloat(oddsSplit[0], 64)
			if err == nil {
				bet.Points = &points
			}
			price, err := strconv.ParseFloat(oddsSplit[1], 64)
			if err == nil {
				bet.Price = price
			}
		}
	}
	sideNode, err := htmlquery.Query(headerNode, `//span[1]`)
	if sideNode != nil {
		bet.SideName = InnerTextTrimmed(sideNode)
	}

	leagueName, err := htmlquery.Query(footerNode, `//div[@class="text-lowlight"][1]`)
	bet.LeagueName = InnerTextTrimmed(leagueName)

	eventTime, err := htmlquery.Query(footerNode, `//div[@class="text-lowlight"][2]`)
	bet.EventTime, err = parseEventTime(InnerTextTrimmed(eventTime))

	sportNode, err := htmlquery.Query(contentNode, `//div[@class="text-lowlight"][1]`)
	teamsNode, err := htmlquery.Query(contentNode, `//div[@class="text-lowlight"][2]`)
	betTimeNode, err := htmlquery.Query(contentNode, `//div[@class="text-lowlight"][5]`)
	bet.BetTime, err = parseEventTime(InnerTextTrimmed(betTimeNode))
	homeNode, err := htmlquery.Query(teamsNode, `//span[1]`)
	awayNode, err := htmlquery.Query(teamsNode, `//span[2]`)
	bet.Home = InnerTextTrimmed(homeNode)
	bet.Away = InnerTextTrimmed(awayNode)

	sportNode, err = htmlquery.Query(contentNode, `//div[@class="text-lowlight"][1]`)
	if sportNode != nil {
		sportMarketSplit := strings.Split(InnerTextTrimmed(sportNode), " / ")
		switch len(sportMarketSplit) {
		case 2:
			bet.SportName = strings.TrimSpace(sportMarketSplit[0])
			bet.MarketName = strings.TrimSpace(sportMarketSplit[1])
		case 1:
			bet.SportName = strings.TrimSpace(sportMarketSplit[0])
		}
	}
	stakeNode, err := htmlquery.Query(contentNode, `//div[@class="info"][1]/div[@class="info_value"]`)
	if stakeNode != nil {
		bet.Stake, err = strconv.ParseFloat(InnerTextTrimmed(stakeNode), 64)
	}
	statusNode, err := htmlquery.QueryAll(contentNode, `//div[@class="info"][2]/div[@class="info_title"]`)
	if statusNode != nil {
		for _, node := range statusNode {
			if InnerTextTrimmed(node) == "Status:" {
				continue
			}
			bet.Status = strings.Trim(InnerTextTrimmed(node), ":")
		}
	}
	wlNode, err := htmlquery.Query(contentNode, `//div[@class="info"][2]/div[@class="info_value"]/span`)
	if wlNode != nil {
		winLoss, err := strconv.ParseFloat(InnerTextTrimmed(wlNode), 64)
		if err == nil {
			bet.WinLoss = &winLoss
		} else if bet.Status == "Refund" {
			bet.WinLoss = util.PtrFloat64(0)
		} else if bet.Status == "Void" {
			bet.WinLoss = util.PtrFloat64(0)
		}
	}
	return &bet, nil
}
func parse(resp string) (*html.Node, error) {
	page, err := html.Parse(strings.NewReader(resp))
	if err != nil {
		return nil, errors.Wrap(err, "parse body error")
	}
	return page, nil
}

var eventTimeRe = regexp.MustCompile(`(\d{4})\/(\d\d)\/(\d\d)\s+(\d\d):(\d\d)`)
var DafabetTimeZone, _ = time.LoadLocation("America/Curacao")

func parseEventTime(eventTime string) (*time.Time, error) {
	etMatch := eventTimeRe.FindStringSubmatch(eventTime)
	if len(etMatch) != 6 {
		return nil, errors.New("parse_event_time_error")
	}
	year, err := strconv.Atoi(etMatch[1])
	if err != nil {
		return nil, err
	}
	month, err := strconv.Atoi(etMatch[2])
	if err != nil {
		return nil, err
	}
	day, err := strconv.Atoi(etMatch[3])
	if err != nil {
		return nil, err
	}
	hour, err := strconv.Atoi(etMatch[4])
	if err != nil {
		return nil, err
	}
	min, err := strconv.Atoi(etMatch[5])
	if err != nil {
		return nil, err
	}
	date := time.Date(year, time.Month(month), day, hour, min, 0, 0, DafabetTimeZone)
	return &date, nil
}
func InnerTextTrimmed(node *html.Node) string {
	text := htmlquery.InnerText(node)
	return strings.TrimSpace(text)
}

func parseBetDetail(detail string, bet *store.Bet) error {
	var d Detail
	bdMatch := BetDetailRe.FindStringSubmatch(detail)
	if len(bdMatch) != 2 {
		return errors.Errorf("bet_detail_parse_error for %v", detail)
	}
	err := json.Unmarshal([]byte(bdMatch[1]), &d)
	if err != nil {
		return err
	}
	bet.SportId = d.SportID
	bet.LeagueId = d.LeagueID
	bet.EventId = d.MatchID
	bet.Id, err = strconv.ParseInt(d.TransID, 10, 64)
	if err != nil {
		return err
	}
	bet.BetTypeId, err = strconv.ParseInt(d.Bettype, 10, 64)
	winLoseDate, err := time.ParseInLocation("2006/01/02", d.WinLoseDate, DafabetTimeZone)
	if err == nil {
		bet.WinLoseDate = &winLoseDate
	}
	return nil
}

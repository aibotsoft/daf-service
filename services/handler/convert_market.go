package handler

import (
	"github.com/aibotsoft/daf-service/pkg/store"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/status"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

var handicapRe = regexp.MustCompile(`Ф(\d)\((-?\d+,?\d{0,3})\)`)
var totalRe = regexp.MustCompile(`Т([МБ])\((-?\d+,?\d{0,3})\)`)
var teamTotalRe = regexp.MustCompile(`ИТ(\d)([МБ])\((-?\d+,?\d{0,3})\)`)
var handicapSide = map[string]string{"1": "h", "2": "a"}

const (
	Handicap        = 1
	OddEven         = 2
	OverUnder       = 3
	FT1X2           = 5
	_1HHandicap     = 7
	_1HOverUnder    = 8
	_1HOddEven      = 12
	_1H1X2          = 15
	_2HHandicap     = 17
	_2HOverUnder    = 18
	Moneyline       = 20
	_1HMoneyline    = 21
	DoubleChance    = 24
	DrawNoBet       = 25
	_1HDoubleChance = 151
	_1HDrawNoBet    = 191
	_2H1X2          = 177
	_2HOddEven      = 184
	_2HDrawNoBet    = 185
	_2HDoubleChance = 186

	HomeTeamOverUnder = 461
	AwayTeamOverUnder = 462

	_HomeTeamOverUnder   = 401
	_AwayTeamOverUnder   = 402
	_1HHomeTeamOverUnder = 403
	_1HAwayTeamOverUnder = 404

	CornersOddEven    = 194
	_1HCornersOddEven = 203
	_2HCornersOddEven = 472

	HomeTeamOverUnderCorners    = 482
	AwayTeamOverUnderCorners    = 483
	_1HHomeTeamOverUnderCorners = 484
	_1HAwayTeamOverUnderCorners = 485

	_1QuarterHandicap  = 609
	_1QuarterOverUnder = 610
	_1QuarterOddEven   = 611
	_1QuarterMoneyline = 612

	QuarterXHomeTeamOverUnder = 615
	QuarterXAwayTeamOverUnder = 616

	MapXMoneyline = 9001
)

//Quarter X Handicap,609
//Quarter X Over/Under,610
//Quarter X Odd/Even,611
//Quarter X Moneyline,612
//	Quarter X Race To Y Points 613

//Quarter X Home Team Over/Under,615
//Quarter X Away Team Over/Under,616

//type Ticket struct {
//	Id        int64
//	BetTeam   string
//	Price     float64
//	BetTypeId int64
//	Points    *float64
//	EventId   int64
//
//	MarketName string
//	IsLive     bool
//	MinBet     float64
//	MaxBet     float64
//	Home       string
//	Away       string
//	IsHandicap bool
//}

//func (t *Ticket) GetPoints() float64 {
//	if t.Points == nil {
//		return 0
//	}
//	return *t.Points
//}
//
//func (t *Ticket) GetSidePoints() float64 {
//	if t.Points == nil {
//		return 0
//	}
//	return *t.Points
//}

//В базе данных гандикап хранится для второй команды, поэтому нужно возвращать его как есть для второй,
//и инвертировать знак для первой команды (только форы)
//func (t *Ticket) GetDbPoints() *float64 {
//	if t.Points == nil {
//		return nil
//	}
//	if !t.IsHandicap {
//		return t.Points
//	}
//	if t.BetTeam == "a" {
//		return t.Points
//	}
//	p := *t.Points * -1
//	return &p
//}

//func calcPoints(side string, points float64) *float64 {
//	if side == "2" {
//		return &points
//	}
//	p := points * -1
//	return &p
//}

var abc = map[string]store.Ticket{
	"П1": {BetTypeId: Moneyline, BetTeam: "h"},
	"П2": {BetTypeId: Moneyline, BetTeam: "a"},

	"(сеты) П1": {BetTypeId: Moneyline, BetTeam: "h"},
	"(сеты) П2": {BetTypeId: Moneyline, BetTeam: "a"},

	"1/2 П1": {BetTypeId: _1HMoneyline, BetTeam: "h"},
	"1/2 П2": {BetTypeId: _1HMoneyline, BetTeam: "a"},

	"2/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 12},
	"2/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 12},
	"3/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 13},
	"3/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 13},
	"4/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 14},
	"4/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 14},
	"5/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 15},
	"5/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 15},
	"6/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 16},
	"6/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 16},
	"7/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 17},
	"7/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 17},
	"8/2 П1": {BetTypeId: MapXMoneyline, BetTeam: "h", Cat: 18},
	"8/2 П2": {BetTypeId: MapXMoneyline, BetTeam: "a", Cat: 18},

	//"1/4 П1": {BetTypeId: "1st Quarter - Money Line", BetTeam: "1"},
	//"1/4 П2": {BetTypeId: "1st Quarter - Money Line", BetTeam: "2"},
	//"2/4 П1": {BetTypeId: "2nd Quarter - Money Line", BetTeam: "1"},
	//"2/4 П2": {BetTypeId: "2nd Quarter - Money Line", BetTeam: "2"},
	//"3/4 П1": {BetTypeId: "3rd Quarter - Money Line", BetTeam: "1"},
	//"3/4 П2": {BetTypeId: "3rd Quarter - Money Line", BetTeam: "2"},
	//"4/4 П1": {BetTypeId: "4th Quarter - Money Line", BetTeam: "1"},
	//"4/4 П2": {BetTypeId: "4th Quarter - Money Line", BetTeam: "2"},

	"Чёт":       {BetTypeId: OddEven, BetTeam: "a"},
	"Нечёт":     {BetTypeId: OddEven, BetTeam: "h"},
	"1/2 Чёт":   {BetTypeId: _1HOddEven, BetTeam: "a"},
	"1/2 Нечёт": {BetTypeId: _1HOddEven, BetTeam: "h"},
	"2/2 Чёт":   {BetTypeId: _2HOddEven, BetTeam: "a"},
	"2/2 Нечёт": {BetTypeId: _2HOddEven, BetTeam: "h"},

	"ЖК Чёт":       {BetTypeId: OddEven, BetTeam: "a"},
	"ЖК Нечёт":     {BetTypeId: OddEven, BetTeam: "h"},
	"ЖК 1/2 Чёт":   {BetTypeId: _1HOddEven, BetTeam: "a"},
	"ЖК 1/2 Нечёт": {BetTypeId: _1HOddEven, BetTeam: "h"},
	"ЖК 2/2 Чёт":   {BetTypeId: _2HOddEven, BetTeam: "a"},
	"ЖК 2/2 Нечёт": {BetTypeId: _2HOddEven, BetTeam: "h"},

	"УГЛ Чёт":       {BetTypeId: CornersOddEven, BetTeam: "E"},
	"УГЛ Нечёт":     {BetTypeId: CornersOddEven, BetTeam: "O"},
	"УГЛ 1/2 Чёт":   {BetTypeId: _1HCornersOddEven, BetTeam: "E"},
	"УГЛ 1/2 Нечёт": {BetTypeId: _1HCornersOddEven, BetTeam: "O"},
	"УГЛ 2/2 Чёт":   {BetTypeId: _2HCornersOddEven, BetTeam: "E"},
	"УГЛ 2/2 Нечёт": {BetTypeId: _2HCornersOddEven, BetTeam: "O"},

	"1/4 Чёт":   {BetTypeId: _1QuarterOddEven, BetTeam: "e"},
	"1/4 Нечёт": {BetTypeId: _1QuarterOddEven, BetTeam: "o"},
	//"2/4 Чёт":   {BetTypeId: "2nd Quarter - Odd/Even", BetTeam: "a"},
	//"2/4 Нечёт": {BetTypeId: "2nd Quarter - Odd/Even", BetTeam: "h"},
	//"3/4 Чёт":   {BetTypeId: "3rd Quarter - Odd/Even", BetTeam: "a"},
	//"3/4 Нечёт": {BetTypeId: "3rd Quarter - Odd/Even", BetTeam: "h"},
	//"4/4 Чёт":   {BetTypeId: "4th Quarter - Odd/Even", BetTeam: "a"},
	//"4/4 Нечёт": {BetTypeId: "4th Quarter - Odd/Even", BetTeam: "h"},

	"1":     {BetTypeId: FT1X2, BetTeam: "1"},
	"2":     {BetTypeId: FT1X2, BetTeam: "2"},
	"Х":     {BetTypeId: FT1X2, BetTeam: "x"},
	"X":     {BetTypeId: FT1X2, BetTeam: "x"},
	"1/2 1": {BetTypeId: _1H1X2, BetTeam: "1"},
	"1/2 2": {BetTypeId: _1H1X2, BetTeam: "2"},
	"1/2 Х": {BetTypeId: _1H1X2, BetTeam: "x"},
	"1/2 X": {BetTypeId: _1H1X2, BetTeam: "x"},
	"2/2 1": {BetTypeId: _2H1X2, BetTeam: "1"},
	"2/2 2": {BetTypeId: _2H1X2, BetTeam: "2"},
	"2/2 Х": {BetTypeId: _2H1X2, BetTeam: "x"},
	"2/2 X": {BetTypeId: _2H1X2, BetTeam: "x"},

	"УГЛ 1":     {BetTypeId: FT1X2, BetTeam: "1"},
	"УГЛ 2":     {BetTypeId: FT1X2, BetTeam: "2"},
	"УГЛ Х":     {BetTypeId: FT1X2, BetTeam: "x"},
	"УГЛ X":     {BetTypeId: FT1X2, BetTeam: "x"},
	"УГЛ 1/2 1": {BetTypeId: _1H1X2, BetTeam: "1"},
	"УГЛ 1/2 2": {BetTypeId: _1H1X2, BetTeam: "2"},
	"УГЛ 1/2 Х": {BetTypeId: _1H1X2, BetTeam: "x"},
	"УГЛ 1/2 X": {BetTypeId: _1H1X2, BetTeam: "x"},
	"УГЛ 2/2 1": {BetTypeId: _2H1X2, BetTeam: "1"},
	"УГЛ 2/2 2": {BetTypeId: _2H1X2, BetTeam: "2"},
	"УГЛ 2/2 Х": {BetTypeId: _2H1X2, BetTeam: "x"},
	"УГЛ 2/2 X": {BetTypeId: _2H1X2, BetTeam: "x"},

	"ЖК 1":     {BetTypeId: FT1X2, BetTeam: "1"},
	"ЖК 2":     {BetTypeId: FT1X2, BetTeam: "2"},
	"ЖК Х":     {BetTypeId: FT1X2, BetTeam: "x"},
	"ЖК X":     {BetTypeId: FT1X2, BetTeam: "x"},
	"ЖК 1/2 1": {BetTypeId: _1H1X2, BetTeam: "1"},
	"ЖК 1/2 2": {BetTypeId: _1H1X2, BetTeam: "2"},
	"ЖК 1/2 Х": {BetTypeId: _1H1X2, BetTeam: "x"},
	"ЖК 1/2 X": {BetTypeId: _1H1X2, BetTeam: "x"},
	"ЖК 2/2 1": {BetTypeId: _2H1X2, BetTeam: "1"},
	"ЖК 2/2 2": {BetTypeId: _2H1X2, BetTeam: "2"},
	"ЖК 2/2 Х": {BetTypeId: _2H1X2, BetTeam: "x"},
	"ЖК 2/2 X": {BetTypeId: _2H1X2, BetTeam: "x"},

	"1Х":     {BetTypeId: DoubleChance, BetTeam: "1x"},
	"1X":     {BetTypeId: DoubleChance, BetTeam: "1x"},
	"Х2":     {BetTypeId: DoubleChance, BetTeam: "2x"},
	"X2":     {BetTypeId: DoubleChance, BetTeam: "2x"},
	"12":     {BetTypeId: DoubleChance, BetTeam: "12"},
	"1/2 1Х": {BetTypeId: _1HDoubleChance, BetTeam: "1x"},
	"1/2 Х2": {BetTypeId: _1HDoubleChance, BetTeam: "2x"},
	"1/2 12": {BetTypeId: _1HDoubleChance, BetTeam: "12"},
	"2/2 1Х": {BetTypeId: _2HDoubleChance, BetTeam: "hd"},
	"2/2 Х2": {BetTypeId: _2HDoubleChance, BetTeam: "da"},
	"2/2 12": {BetTypeId: _2HDoubleChance, BetTeam: "ha"},

	"ЖК 1Х":     {BetTypeId: DoubleChance, BetTeam: "1x"},
	"ЖК Х2":     {BetTypeId: DoubleChance, BetTeam: "2x"},
	"ЖК 12":     {BetTypeId: DoubleChance, BetTeam: "12"},
	"ЖК 1/2 1Х": {BetTypeId: _1HDoubleChance, BetTeam: "1x"},
	"ЖК 1/2 Х2": {BetTypeId: _1HDoubleChance, BetTeam: "2x"},
	"ЖК 1/2 12": {BetTypeId: _1HDoubleChance, BetTeam: "12"},
	"ЖК 2/2 1Х": {BetTypeId: _2HDoubleChance, BetTeam: "hd"},
	"ЖК 2/2 Х2": {BetTypeId: _2HDoubleChance, BetTeam: "da"},
	"ЖК 2/2 12": {BetTypeId: _2HDoubleChance, BetTeam: "ha"},

	"УГЛ 1Х":     {BetTypeId: DoubleChance, BetTeam: "1x"},
	"УГЛ Х2":     {BetTypeId: DoubleChance, BetTeam: "2x"},
	"УГЛ 12":     {BetTypeId: DoubleChance, BetTeam: "12"},
	"УГЛ 1/2 1Х": {BetTypeId: _1HDoubleChance, BetTeam: "1x"},
	"УГЛ 1/2 Х2": {BetTypeId: _1HDoubleChance, BetTeam: "2x"},
	"УГЛ 1/2 12": {BetTypeId: _1HDoubleChance, BetTeam: "12"},
	"УГЛ 2/2 1Х": {BetTypeId: _2HDoubleChance, BetTeam: "hd"},
	"УГЛ 2/2 Х2": {BetTypeId: _2HDoubleChance, BetTeam: "da"},
	"УГЛ 2/2 12": {BetTypeId: _2HDoubleChance, BetTeam: "ha"},

	"Старт К1": {BetTypeId: Handicap, BetTeam: "h"},
	"Старт К2": {BetTypeId: Handicap, BetTeam: "a"},
}

func processPoint(point string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(point, ",", ".", -1), 64)
}

var urlRe = regexp.MustCompile(`matchid=(\d+)&eventState=(\w+)&sport=(\d+)&leagueID=(-?\d+)`)
var eventStateMap = map[string]string{"today": "t", "earlyMarket": "e", "live": "l"}

func (h *Handler) ParseUrl(side *pb.SurebetSide, line *store.Ticket) (err error) {
	side.Check.Status = status.StatusError
	side.Check.StatusInfo = "parse_url_error"
	u := urlRe.FindStringSubmatch(side.Url)
	if len(u) < 5 {
		h.log.Infow(side.Check.StatusInfo, "url", side.Url)
		return CheckLineError
	}
	if side.EventId == "" {
		side.EventId = u[1]
	}
	if side.SportId == 0 {
		side.SportId, err = strconv.ParseInt(u[3], 10, 64)
		if err != nil {
			return CheckLineError
		}
	}
	if side.LeagueId == 0 {
		side.LeagueId, err = strconv.ParseInt(u[4], 10, 64)
		if err != nil {
			return CheckLineError
		}
	}
	line.EventState = eventStateMap[u[2]]
	//h.log.Infow("parse_url_ok", "eventId", side.EventId, "sportId", side.SportId, "leagueId", side.LeagueId, "eventState", line.EventState)
	return nil
}
func Convert(side *pb.SurebetSide) (*store.Ticket, error) {
	name := side.MarketName
	if name == "" {
		return nil, errors.Errorf("cannot convert empty str")
	}
	m := abc[name]
	if m.BetTypeId != 0 {
		if side.SportId == 43 {
			if strings.Index(name, "1/2 П") != -1 {
				m.Cat = 11
				m.BetTypeId = MapXMoneyline
			} else if strings.Index(name, "2/2 П") != -1 {
				m.Cat = 12
				m.BetTypeId = MapXMoneyline
			}
		}
		return &m, nil
	}
	var found []string
	found = handicapRe.FindStringSubmatch(name)
	if len(found) == 3 {
		return processHandicap(name, found)
	}
	found = totalRe.FindStringSubmatch(name)
	if len(found) == 3 {
		return processTotal(name, found)
	}
	found = teamTotalRe.FindStringSubmatch(name)
	if len(found) == 4 {
		return processTeamTotal(name, found, side.SportId)
	}
	return nil, errors.Errorf("not converted market: %s", name)
}
func processHandicap(market string, found []string) (*store.Ticket, error) {
	point, err := processPoint(found[2])
	if err != nil {
		return nil, err
	}
	var typeId int64 = Handicap
	switch {
	case strings.Index(market, "1/2") != -1:
		typeId = _1HHandicap
	case strings.Index(market, "2/2") != -1:
		typeId = _2HHandicap
	case strings.Index(market, "1/4") != -1:
		typeId = _1QuarterHandicap
		//case strings.Index(market, "2/4") != -1:
		//	typeId = "2nd Quarter - Handicap"
		//case strings.Index(market, "3/4") != -1:
		//	typeId = "3rd Quarter - Handicap"
		//case strings.Index(market, "4/4") != -1:
		//	typeId = "4th Quarter - Handicap"
	}
	m := store.Ticket{BetTypeId: typeId, BetTeam: handicapSide[found[1]], Points: &point, IsHandicap: true}
	return &m, nil
}

var totalSide = map[string]string{"Б": "h", "М": "a"}

func processTotal(market string, found []string) (*store.Ticket, error) {
	point, err := processPoint(found[2])
	if err != nil {
		return nil, err
	}
	var typeId int64 = OverUnder
	switch {
	case strings.Index(market, "1/2") != -1:
		typeId = _1HOverUnder
	case strings.Index(market, "2/2") != -1:
		typeId = _2HOverUnder
	case strings.Index(market, "1/4") != -1:
		typeId = _1QuarterOverUnder
		//case strings.Index(market, "2/4") != -1:
		//	typeId = "2nd Quarter - Over/Under"
		//case strings.Index(market, "3/4") != -1:
		//	typeId = "3rd Quarter - Over/Under"
		//case strings.Index(market, "4/4") != -1:
		//	typeId = "4th Quarter - Over/Under"
	}
	m := store.Ticket{BetTypeId: typeId, BetTeam: calcTotalSide(typeId, found[1]), Points: &point}
	return &m, nil
}
func calcTotalSide(typeId int64, side string) string {
	if typeId <= 8 {
		return totalSide[side]
	}
	return teamTotalSide[side]
}

var teamTotalSide = map[string]string{"Б": "O", "М": "U"}

func processTeamTotal(market string, found []string, sportId int64) (*store.Ticket, error) {
	point, err := processPoint(found[3])
	if err != nil {
		return nil, err
	}
	var typeId int64
	if found[1] == "1" {
		typeId = HomeTeamOverUnder
		switch {
		case strings.Index(market, "УГЛ ИТ1") != -1:
			typeId = HomeTeamOverUnderCorners
		case strings.Index(market, "УГЛ 1/2 ИТ1") != -1:
			typeId = _1HHomeTeamOverUnderCorners
		case strings.Index(market, "1/2") != -1:
			typeId = _1HHomeTeamOverUnder
		case strings.Index(market, "1/4") != -1:
			typeId = QuarterXHomeTeamOverUnder
		case sportId == 2:
			typeId = _HomeTeamOverUnder

		}
	} else {
		typeId = AwayTeamOverUnder
		switch {
		case strings.Index(market, "УГЛ ИТ2") != -1:
			typeId = AwayTeamOverUnderCorners
		case strings.Index(market, "УГЛ 1/2 ИТ2") != -1:
			typeId = _1HAwayTeamOverUnderCorners
		case strings.Index(market, "1/2") != -1:
			typeId = _1HAwayTeamOverUnder
		case strings.Index(market, "1/4") != -1:
			typeId = QuarterXAwayTeamOverUnder
		case sportId == 2:
			typeId = _AwayTeamOverUnder
		}
	}
	m := store.Ticket{BetTypeId: typeId, BetTeam: teamTotalSide[found[2]], Points: &point}
	return &m, nil
}

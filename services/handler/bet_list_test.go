package handler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var betsResponse = `
    <div class="bets bets-" style="display: block;">
            <div class="bets_header">
                <div class="bets_title">
                    <span class="">FT.1</span>
                    
                        <span class="text-odds">@2.12</span>
                </div>
            </div>

        <div class="bets_content touch" onclick="RunningCollapseDetailClick(this)">
                    <div class="text-lowlight">Soccer / FT.1X2</div>
                    <div class="text-lowlight">
                        
                        <span>Australia (V)</span> vs <span>Costa Rica (V)</span>
                    </div>
                    <div class="text-lowlight"></div>
                    <div class="text-lowlight"></div>




				<div class="text-lowlight">Bet Time : 2020/05/17 23:17</div>

            <div class="bets_total">
                <div class="info">
                    <div class="info_title">Stake:</div>
                        <div class="info_value">1</div>

                </div>
                <div class="info">


                        <div class="info_title">Won:</div>
                            <div class="info_value"><span class="text-win">1.12</span></div>

                </div>
                <i class="icon icon-arrow-bottom"></i>
            </div>
        </div>




        <div class="bets_footer">
                <div class="text-lowlight">SABA INTERNATIONAL FRIENDLY Virtual PES 20 - 20 Mins Play</div>
                                <div class="text-lowlight">Event Time : 2020/05/18 01:30 </div>
                    <div class="text-lowlight bets_moredetail">
                        Result:
                        <div class="btn btn-link" href="javascript:void(null);" onclick='ShowPopupResult([{"SportID":1,"LeagueID":89692,"MatchID":36137514,"MatchIDString":null,"Bettype":"5","BetID":"0","TransID":"109054039552","WinLoseDate":"2020/05/17","TransDesc":null,"TransID3rd":null,"VendorID":0,"IsSuperLive":false,"HDP":"1","MarketID":null}])'>
                            <div class="btn_text">Detail</div>
                            <i class="icon icon-arrow-right"></i>
                        </div>
                    </div>
                    <div class="text-lowlight">Odds Type: Dec</div>


			
            
            <div class="text-lowlight">ID: 109054039552 </div>
            <div class="bets_info">
            </div>
        </div>
    </div>
    <div class="bets bets-" style="display: block;">
            <div class="bets_header">
                <div class="bets_title">
                    <span class="">FT.1</span>
                    
                        <span class="text-odds">@2.12</span>
                </div>
            </div>

        <div class="bets_content touch" onclick="RunningCollapseDetailClick(this)">
                    <div class="text-lowlight">Soccer / FT.1X2</div>
                    <div class="text-lowlight">
                        
                        <span>Australia (V)</span> vs <span>Costa Rica (V)</span>
                    </div>
                    <div class="text-lowlight"></div>
                    <div class="text-lowlight"></div>




				<div class="text-lowlight">Bet Time : 2020/05/17 23:17</div>

            <div class="bets_total">
                <div class="info">
                    <div class="info_title">Stake:</div>
                        <div class="info_value">1</div>

                </div>
                <div class="info">


                        <div class="info_title">Won:</div>
                            <div class="info_value"><span class="text-win">1.12</span></div>

                </div>
                <i class="icon icon-arrow-bottom"></i>
            </div>
        </div>




        <div class="bets_footer">
                <div class="text-lowlight">SABA INTERNATIONAL FRIENDLY Virtual PES 20 - 20 Mins Play</div>
                                <div class="text-lowlight">Event Time : 2020/05/18 01:30 </div>
                    <div class="text-lowlight bets_moredetail">
                        Result:
                        <div class="btn btn-link" href="javascript:void(null);" onclick='ShowPopupResult([{"SportID":1,"LeagueID":89692,"MatchID":36137514,"MatchIDString":null,"Bettype":"5","BetID":"0","TransID":"109054040611","WinLoseDate":"2020/05/17","TransDesc":null,"TransID3rd":null,"VendorID":0,"IsSuperLive":false,"HDP":"1","MarketID":null}])'>
                            <div class="btn_text">Detail</div>
                            <i class="icon icon-arrow-right"></i>
                        </div>
                    </div>
                    <div class="text-lowlight">Odds Type: Dec</div>


			
            
            <div class="text-lowlight">ID: 109054040611 </div>
            <div class="bets_info">
            </div>
        </div>
    </div>
    <div class="bets bets-" style="display: block;">
            <div class="bets_header">
                <div class="bets_title">
                    <span class="">FT.1</span>
                    
                        <span class="text-odds">@2.12</span>
                </div>
            </div>

        <div class="bets_content touch" onclick="RunningCollapseDetailClick(this)">
                    <div class="text-lowlight">Soccer / FT.1X2</div>
                    <div class="text-lowlight">
                        
                        <span>Australia (V)</span> vs <span>Costa Rica (V)</span>
                    </div>
                    <div class="text-lowlight"></div>
                    <div class="text-lowlight"></div>




				<div class="text-lowlight">Bet Time : 2020/05/17 23:13</div>

            <div class="bets_total">
                <div class="info">
                    <div class="info_title">Stake:</div>
                        <div class="info_value">1</div>

                </div>
                <div class="info">


                        <div class="info_title">Won:</div>
                            <div class="info_value"><span class="text-win">1.12</span></div>

                </div>
                <i class="icon icon-arrow-bottom"></i>
            </div>
        </div>




        <div class="bets_footer">
                <div class="text-lowlight">SABA INTERNATIONAL FRIENDLY Virtual PES 20 - 20 Mins Play</div>
                                <div class="text-lowlight">Event Time : 2020/05/18 01:30 </div>
                    <div class="text-lowlight bets_moredetail">
                        Result:
                        <div class="btn btn-link" href="javascript:void(null);" onclick='ShowPopupResult([{"SportID":1,"LeagueID":89692,"MatchID":36137514,"MatchIDString":null,"Bettype":"5","BetID":"0","TransID":"109054035184","WinLoseDate":"2020/05/17","TransDesc":null,"TransID3rd":null,"VendorID":0,"IsSuperLive":false,"HDP":"1","MarketID":null}])'>
                            <div class="btn_text">Detail</div>
                            <i class="icon icon-arrow-right"></i>
                        </div>
                    </div>
                    <div class="text-lowlight">Odds Type: Dec</div>


			
            
            <div class="text-lowlight">ID: 109054035184 </div>
            <div class="bets_info">
            </div>
        </div>
    </div>
<div class="bets bets-running" style="display: block;">
            <div class="bets_header">
                <div class="bets_title">
                    <span class="">GAM Esports</span>
                    
                        <span class="text-odds">@3.08</span>
                </div>
            </div>

        <div class="bets_content touch" onclick="RunningCollapseDetailClick(this)">
                    <div class="text-lowlight">E-Sports /  LOL -  Map 3 Moneyline</div>
                    <div class="text-lowlight">
                        
                        <span>Team Flash</span> vs <span>GAM Esports</span>
                    </div>
                    <div class="text-lowlight"></div>
                    <div class="text-lowlight"></div>




				<div class="text-lowlight">Bet Time : 2020/05/23 04:56</div>

            <div class="bets_total">
                <div class="info">
                    <div class="info_title">Stake:</div>
                        <div class="info_value">38</div>

                </div>
                <div class="info">
                        <div class="info_title">Status:</div>


                        <div class="info_title">Refund</div>
                        <div class="info_value"><span class=""></span></div>

                </div>
                <i class="icon icon-arrow-bottom"></i>
            </div>
        </div>




        <div class="bets_footer">
                <div class="text-lowlight">League of Legends - Pulsefire Cup</div>
                                <div class="text-lowlight">Event Time : 2020/05/23 05:00 </div>
                    <div class="text-lowlight bets_moredetail">
                        Result:
                        <div class="btn btn-link" href="javascript:void(null);" onclick='ShowPopupResult([{"SportID":43,"LeagueID":90566,"MatchID":36203174,"MatchIDString":null,"Bettype":"9001","BetID":"03","TransID":"109083142151","WinLoseDate":"2020/05/23","TransDesc":null,"TransID3rd":null,"VendorID":0,"IsSuperLive":false,"HDP":"0","MarketID":null}])'>
                            <div class="btn_text">Detail</div>
                            <i class="icon icon-arrow-right"></i>
                        </div>
                    </div>
                    <div class="text-lowlight">Odds Type: Dec</div>


			
            
            <div class="text-lowlight">ID: 109083142151 </div>
            <div class="bets_info">
            </div>
        </div>
    </div>

`

func Test_processBets(t *testing.T) {
	got, err := h.processBets(betsResponse)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, got)
		//t.Log(got)
	}
}
func TestHandler_BetList(t *testing.T) {
	err := h.BetList(context.Background())
	assert.NoError(t, err)

}

create or alter proc dbo.uspCalcStat @EventId int
as
begin
    set nocount on;
    select s.MarketName,
           count(s.EventId) over ( )                                   CountEvent,
           cast(sum(b.Stake) over ( ) as int)                          AmountEvent,
           count(s.MarketName) over ( partition by s.MarketName)       CountLine,
           cast(sum(b.Stake) over ( partition by s.MarketName) as int) AmountLine
    from dbo.Bet b
             join dbo.Side s on s.Id = b.SurebetId and s.SideIndex = b.SideIndex
    where EventId = @EventId and b.Status = 'Ok'
end;

--     exec dbo.uspCalcStat 38236261, 'ТМ(3)'

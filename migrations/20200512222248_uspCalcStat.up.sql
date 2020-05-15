create or alter proc dbo.uspCalcStat @EventId int, @MarketName varchar(100)
as
begin
    set nocount on;
    with cte as (
        select s.MarketName,
               count(s.EventId) over ( )                                   CountEvent,
               cast(sum(b.Stake) over ( ) as int)                          AmountEvent,
               count(s.MarketName) over ( partition by s.MarketName)       CountLine,
               cast(sum(b.Stake) over ( partition by s.MarketName) as int) AmountLine
        from dbo.Bet b
                 join dbo.Side s on s.Id = b.SurebetId
        where EventId = @EventId and b.Status = 'Ok'
    )
    select CountEvent, AmountEvent, CountLine, AmountLine
    from cte
    where cte.MarketName = @MarketName
end
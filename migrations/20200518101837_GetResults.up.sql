create or alter view dbo.GetResults as
select top 3000 b.SurebetId,
       b.SideIndex,
       b.BetId,
       b.ApiBetId,
       l.Status  ApiBetStatus,
       l.Price   Price,
       l.Stake   Stake,
       l.WinLoss WinLoss
from Bet b
         left join dbo.BetList l on b.ApiBetId = l.Id
where b.ApiBetId > 0 and l.Status is not null
order by SurebetId desc


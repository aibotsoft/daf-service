create or alter function dbo.fnGetDemoEvents(@Cnt int, @SportId int, @HomeNameLike varchar(200))
    returns table
        return
        with t as (
            select top (@Cnt) l.EventId
            from Event e
                     join Line l on l.EventId = e.Id
                     join Team h on h.Id = e.Home

            where e.SportId = @SportId
              and e.Starts > sysdatetimeoffset()
              and h.Name like @HomeNameLike
            group by l.EventId
            order by count(l.Id) desc
        )
        select e.Id EventId,
               s.Id SportId,
               s.Name SportName,
               l.Name LeagueName,
               l.Id LeagueId,
               h.Name Home,
               a.Name Away,
               e.EventState EventState
        from t
                 join Event e on t.EventId = e.Id
                 join Sport s on s.Id = e.SportId
                 join League l on l.Id = e.LeagueId
                 join Team h on e.Home = h.Id
                 join Team a on e.Away = a.Id


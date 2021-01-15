create or alter proc dbo.uspCreateBetList @TVP dbo.BetListType READONLY as
begin
    set nocount on

    MERGE dbo.BetList AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET SportName   = s.SportName,
            SportId     = s.SportId,
            LeagueName  = s.LeagueName,
            LeagueId    = s.LeagueId,
            Home        = s.Home,
            Away        = s.Away,
            EventId     = s.EventId,
            MarketName  = s.MarketName,
            SideName  = s.SideName,
            Points      = s.Points,
            EventTime   = s.EventTime,
            Status      = s.Status,
            WinLoss     = s.WinLoss,
            Price       = s.Price,
            Stake       = s.Stake,
            BetTypeId   = s.BetTypeId,
            BetTime     = s.BetTime,
            WinLoseDate = s.WinLoseDate,
            UpdatedAt   =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, SportName, SportId, LeagueName, LeagueId, Home, Away, EventId, MarketName, SideName, Points, EventTime,
                Status, WinLoss, Price, Stake, BetTypeId, BetTime, WinLoseDate)
        VALUES (s.Id, s.SportName, s.SportId, s.LeagueName, s.LeagueId, s.Home, s.Away, s.EventId, s.MarketName, s.SideName,
                s.Points, s.EventTime, s.Status, s.WinLoss, s.Price, s.Stake, s.BetTypeId, s.BetTime, s.WinLoseDate);
end
create or alter proc dbo.uspCreateEvents @TVP dbo.EventType READONLY as
begin
    set nocount on

    MERGE dbo.Event AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET Home       = s.Home,
            Away       = s.Away,
            LeagueId   = s.LeagueId,
            SportId    = s.SportId,
            EventState = s.EventState,
            UpdatedAt  = sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, Home, Away, LeagueId, SportId, EventState)
        VALUES (s.Id, s.Home, s.Away, s.LeagueId, s.SportId, s.EventState);
end
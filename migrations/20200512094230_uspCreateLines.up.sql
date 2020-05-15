create or alter proc dbo.uspCreateLines @TVP dbo.LineType READONLY as
begin
    set nocount on

    MERGE dbo.Line AS t
    USING @TVP s
    ON (t.Id = s.Id and t.BetTeam = s.BetTeam)

    WHEN MATCHED THEN
        UPDATE
        SET Price     = s.Price,
            BetTypeId=s.BetTypeId,
            Points=s.Points,
            EventId=s.EventId,
            UpdatedAt =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, BetTeam, Price, BetTypeId, Points, EventId)
        VALUES (s.Id, s.BetTeam, s.Price, s.BetTypeId, s.Points, EventId);
end
create or alter proc dbo.uspCreateLines @TVP dbo.LineType READONLY as
begin
    set nocount on

    MERGE dbo.Line AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET BetTypeId = s.BetTypeId,
            Points = s.Points,
            EventId = s.EventId,
            Cat = s.Cat,
            UpdatedAt =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id,  BetTypeId, Points, EventId, Cat)
        VALUES (s.Id,  s.BetTypeId, s.Points, EventId, s.Cat);
end
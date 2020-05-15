create or alter proc dbo.uspCreateLeagues @TVP dbo.LeagueType READONLY as
begin
    set nocount on

    MERGE dbo.League AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET Name      = s.Name,
            SportId   = s.SportId,
            UpdatedAt =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, Name, SportId)
        VALUES (s.Id, s.Name, s.SportId);
end
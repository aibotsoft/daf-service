create or alter proc dbo.uspCreateTeams @TVP dbo.TeamType READONLY as
begin
    set nocount on

    MERGE dbo.Team AS t
    USING @TVP s
    ON (t.Id = s.Id and t.SportId = s.SportId)

    WHEN MATCHED THEN
        UPDATE
        SET Name      = s.Name,
            UpdatedAt = sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, Name, SportId)
        VALUES (s.Id, s.Name, s.SportId);
end
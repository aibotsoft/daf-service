create or alter proc dbo.uspCreateSports @TVP dbo.SportType READONLY as
begin
    set nocount on

    MERGE dbo.Sport AS t
    USING @TVP s
    ON (t.Id = s.Id)

    WHEN MATCHED THEN
        UPDATE
        SET Name      = s.Name,
            UpdatedAt = sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Id, Name)
        VALUES (s.Id, s.Name);
end
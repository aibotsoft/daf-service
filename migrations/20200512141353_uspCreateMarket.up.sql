create or alter proc dbo.uspCreateMarkets @TVP dbo.MarketType READONLY as
begin
    set nocount on

    MERGE dbo.Market AS t
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
create proc dbo.uspSaveSession @Session varchar(50), @Token varchar(50)
as
begin
    set nocount on
    MERGE dbo.Auth AS t
    USING (select @Session) s (Session)
    on t.Session = s.Session

    WHEN MATCHED THEN
        UPDATE
        SET LastCheckAt = sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (Session, Token) VALUES (@Session, @Token);
end

create or alter proc dbo.uspCreateMarket @Id int, @Name varchar(300) as
begin
    set nocount on
    declare @OldId int

    select @OldId = Id from dbo.Market where Id = @Id
    if @@rowcount = 0
        insert into dbo.Market (Id, Name) values (@Id, @Name)
    else
        select @OldId
end
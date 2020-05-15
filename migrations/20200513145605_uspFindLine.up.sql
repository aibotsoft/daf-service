create or alter proc dbo.uspFindLine @BetTeam varchar(50), @BetTypeId int, @EventId int, @Points decimal(9, 6) as
begin
    set nocount on
    if @Points is null
        select Id
--                Price
        from dbo.Line
        where EventId = @EventId
          and BetTypeId = @BetTypeId
          and BetTeam = @BetTeam
          and Points is null
    else
        select Id
--                Price
        from dbo.Line
        where EventId = @EventId
          and BetTypeId = @BetTypeId
          and BetTeam = @BetTeam
          and Points = @Points
end


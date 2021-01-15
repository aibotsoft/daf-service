create or alter proc dbo.uspFindLine @BetTypeId int, @EventId int, @Points decimal(9, 6), @Cat int as
begin
    set nocount on
    if @Points is null
        select Id
        from dbo.Line
        where EventId = @EventId
          and BetTypeId = @BetTypeId
            and Cat = @Cat
    else
        select Id
        from dbo.Line
        where EventId = @EventId
          and BetTypeId = @BetTypeId
            and Cat = @Cat
          and Points = @Points
end


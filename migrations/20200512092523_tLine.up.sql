create table dbo.Line
(
    Id        int                                        not null,
    BetTypeId int                                        not null,
    Points    decimal(9, 6),
    Cat       int                                        not null,
    EventId   int                                        not null,
    CreatedAt datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_Line primary key (Id),
    index IX_Line (EventId, BetTypeId, Points)
);
create type dbo.LineType as table
(
    Id         int         not null,
    BetTeam    varchar(50) not null,
    Price      decimal(9, 5),
    BetTypeId  int         not null,
    Points     decimal(9, 6),
    EventId    int         not null,

    MarketName varchar(1000),
    IsLive     bit,
    MinBet     decimal(9, 5),
    MaxBet     decimal(9, 5),
    Home       varchar(1000),
    Away       varchar(1000),
    Cat        int,
    IsHandicap        bit,
    EventState        varchar(1000),
    primary key (Id)
)
create table dbo.Sport
(
    Id        int                                        not null,
    Name      varchar(180)                               not null,
    CreatedAt datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_SportId primary key (Id),
)
create type dbo.SportType as table
(
    Id     int          not null,
    Name   varchar(180) not null,
    Count  int,
    EventState varchar(180),
    primary key (Id)
)
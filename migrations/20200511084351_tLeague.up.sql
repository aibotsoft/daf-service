create table dbo.League
(
    Id        int                                        not null,
    Name      varchar(1000)                              not null,
    SportId   int                                        not null,
    CreatedAt datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_LeagueId primary key (Id),
)

create type dbo.LeagueType as table
(
    Id         int           not null,
    Name       varchar(1000) not null,
    SportId    int           not null,
    EventState varchar(300)  not null,
    primary key (Id)
)
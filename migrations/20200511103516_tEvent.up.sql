create table dbo.Event
(
    Id         int                                        not null,
    Home       varchar(300)                               not null,
    Away       varchar(300)                               not null,

    LeagueId   int                                        not null,
    SportId    int                                        not null,
    EventState varchar(300)                               not null,
    Starts     datetimeoffset,
    CreatedAt  datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt  datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_Event primary key (Id),
);
create type dbo.EventType as table
(
    Id         int          not null,
    Home       varchar(300) not null,
    Away       varchar(300) not null,
    LeagueId   int          not null,
    SportId    int          not null,
    EventState varchar(300) not null,
    primary key (Id)
)
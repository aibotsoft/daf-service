create table dbo.Team
(
    Id        int                                        not null,
    Name      varchar(2000)                              not null,
    SportId   int                                        not null,
    CreatedAt datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_Team primary key (Id, SportId),
)

create type dbo.TeamType as table
(
    Id         int           not null,
    Name       varchar(2000) not null,
    SportId    int           not null,
    primary key (Id, SportId)
)
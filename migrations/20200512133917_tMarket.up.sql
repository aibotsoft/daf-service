create table dbo.Market
(
    Id        int                                        not null,
    Name      varchar(300)                               not null,
    CreatedAt datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_Market primary key (Id),
)
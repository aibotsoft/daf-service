create table dbo.Auth
(
--     Id          int identity                               not null,
    Session     varchar(50)                                not null,
    Token       varchar(50)                                not null,
    CreatedAt   datetimeoffset default sysdatetimeoffset() not null,
    LastCheckAt datetimeoffset default sysdatetimeoffset() not null,
    constraint PK_Auth primary key (Session),
)

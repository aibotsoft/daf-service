create table dbo.Bet
(
    SurebetId  bigint                                     not null,
    SideIndex  tinyint                                    not null,

    BetId      bigint,
    TryCount   tinyint,

    Status     varchar(1000),
    StatusInfo varchar(1000),
    Start      bigint,
    Done       bigint,

    Price      decimal(9, 5),
    Stake      decimal(9, 5),
    ApiBetId   bigint,

    CreatedAt  datetimeoffset default sysdatetimeoffset() not null,
    UpdatedAt  datetimeoffset default sysdatetimeoffset() not null,

    constraint PK_Bet primary key (SurebetId, SideIndex),
)
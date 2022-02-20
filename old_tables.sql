CREATE TABLE Cards (
    CardID uuid NOT NULL PRIMARY KEY,
    ScryfallUri varchar(512),
    CardName varchar(255),
    Color varchar(255),
    ColorIdentity varchar(255),
    Type varchar(255),
    CMC double precision,
    ManaCost varchar(255),
    OracleText varchar(1024)
);

CREATE TABLE Players (
    PlayerID uuid NOT NULL PRIMARY KEY,
    Name varchar(30) NOT NULL,
    DiscordID bigint
);

CREATE TABLE Tournaments (
    TournamentID uuid NOT NULL PRIMARY KEY,
    Format varchar(25),
    Location varchar(25),
    Structure varchar(25) NOT NULL,
    Date timestamp,
    TournamentName varchar(1024) NOT NULL
);

CREATE TABLE TournamentPlayers (
    TournamentID uuid NOT NULL references Tournaments(TournamentID),
    PlayerID uuid NOT NULL references Players(PlayerID),
    TriceName varchar(255)
);

CREATE TABLE Decks (
    DeckID uuid NOT NULL PRIMARY KEY,
    DeckHash char(8)
);

CREATE TABLE Commanders (
    DeckID uuid NOT NULL references Decks(DeckID),
    CardID uuid NOT NULL references Cards(CardID)
);

CREATE TABLE DeckCards (
    DeckID uuid NOT NULL references Decks(DeckID),
    CardID uuid NOT NULL references Cards(CardID),
    Count int,
    Sideboard boolean
);

CREATE TABLE TournamentDecks (
    DeckID uuid NOT NULL references Decks(DeckID),
    PlayerID uuid NOT NULL references Players(PlayerID),
    TournamentID uuid NOT NULL references Tournaments(TournamentID),
    DeckName varchar(30) NOT NULL
);

CREATE TABLE Matches (
    MatchID uuid NOT NULL PRIMARY KEY,
    TournamentID uuid NOT NULL references Tournaments(TournamentID),
    WinnerID uuid references Players(PlayerID),
    ReplayURL varchar(255),
    Turns int,
    Spectators int,
    StartTime timestamp,
    EndTime timestamp,
    TimeExtension int,
    MatchNumber int,
    TriceMatch boolean
);

CREATE TABLE MatchPlayers (
    PlayerID uuid NOT NULL references Players(PlayerID),
    MatchID uuid NOT NULL references Matches(MatchID),
    DeckID uuid references Decks(DeckID)
);

CREATE INDEX ON players(name);
CREATE INDEX ON players(discordid);
CREATE INDEX ON cards(cardname);
CREATE INDEX ON tournaments(format);
CREATE INDEX ON decks(deckhash);
CREATE INDEX ON deckcards(cardid);
CREATE INDEX ON tournamentdecks(deckname);
CREATE INDEX ON tournamentdecks(playerid);
CREATE INDEX ON tournamentdecks(tournamentid);
CREATE INDEX ON matches(ReplayURL);
CREATE INDEX ON matches(tournamentid);
CREATE INDEX ON matches(winnerid);
CREATE INDEX ON matchplayers(playerid);
CREATE INDEX ON matchplayers(matchid);
CREATE INDEX ON matchplayers(deckid);
CREATE INDEX ON commanders(cardid);
CREATE INDEX ON commanders(deckid);


/* Card cache */
craete table Cards (
    oracle_id uuid not null primary key,
    scryfall_uri varchar(512),
    card_name varchar(255),
    color varchar(255),
    color_identity varchar(255),
    type varchar(255),
    cmc double precision,
    mana_cost varchar(255),
    oracle_text varchar(1024)
);

/* Squire ID Model */

create table players (
  uuid player_id primary key; /* An internal id for the player */
  varchar(255) email not null;
  varchar(255) hashed_pwd not null;
  varchar(255) pwd_salt not null;

  /* Squire Core Settings */
  varchar(255) player_name not null unique;
  varchar(255) default_game_name;

  /* Third Party Integrations (for identification from unregistered users + account linking) */
  varchar(255) dci_number; /* this is still a thing right? */
  varchar(255) mtga_username;
  bigint discord_id; /* discord integration */
);

/* Squire Core Model */

create table tournament_settings (
  uuid settings_id primary key;
	boolean make_vc not null;
  boolean make_tc not null;
  boolean trice_bot not null;
  boolean spectators_allowed not null;
  boolean spectators_can_see_hands not null;
  boolean spectators_can_chat not null;
  boolean only_registered not null;

  varchar(50) format not null;
  int match_duration check(tournament_settings.match_duration > 0) not null;
  int match_players check(tournament_settings.match_players > 0) not null;

  /* No constraint as some people may have very funny match making */  
  int win_points not null;
  int lose_points not null;
  int draw_points not null;
);

create table tournaments (
	uuid tourn_id primary key;
  uuid settings_id references tournament_settings(settings_id) not null;
  timestamp create_time not null;
  timestamp end_time;
  boolean registration_open;
);

create table decks (
	uuid deck_id primary key;
  uuid player_id references players(player_id) not null;
  /* deck hash is derived so ommitted */
);

create table deck_cards (
	uuid deck_id references decks(deck_id) not null;
  uuid oracle_id not null;
);

create table matches (
	uuid match_id primary key;
  uuid tourn_id references tournaments(tourn_id) not null;
  timestamp create_time;
  timestamp finish_time;
  boolean cancelled;
);

create table match_players(
	uuid match_id references matches(match_id) not null;
  uuid player_id references players(player_id) not null;
  int player_status; /* This is a software defined enum, 0 = lose, 1 = win, 2 = draw */  
);

/* Squire Bot Model */

create table guilds (
	bigint guild_id primary key;
  varchar(255) guild_name not null; /* A cache of the guild name */
  default_settings_id references tournament_settings(settings_id) not null;
);

create table discord_tournaments (
	uuid toun_id references tournaments(toun_id) not null;
  uuid guild_id references guilds(guild_id) not null;
  unique(tourn_id, guild_id);
);

create table discord_players(				
	uuid player_id references players(player_id) not null;
  bigint discord_id not null;
  unique(player_id, discord_id);
);


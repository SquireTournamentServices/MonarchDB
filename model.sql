/* Card cache */
create table types (
	type_filtered varchar(255) primary key,
  type varchar(255)
);

create table cards (
  oracle_id uuid primary key,
  scryfall_uri varchar(512) not null,
  card_name varchar(255) not null unique,
  filtered_name varchar(255) not null,
  cmc double precision not null,
  mana_cost varchar(255) not null,
  oracle_text varchar(1024) not null
);

create table card_types (
	oracle_id uuid not null references(cards.oracle_id),
  type_filtered varchar(255) not null references(types.type_filtered),
  unique(oracle_id, type_filtered)
);

create table card_colors {
	oracle_id uuid not null references(cards.oracle_id),
	color varchar(5) not null
};

create table card_identity_colors {
	oracle_id uuid not null references(cards.oracle_id),
	color varchar(5) not null
};

create table sets (
	set_id uuid primary key,
  set_name varchar(255) not null,
  filtered_name varchar(255) not null,
  set_icon_uri varchar(255) not null,
  set_release timestamp not null
);

create table card_sets (
	oracle_id references cards(oracle_id) not null,
  set_id references sets(set_id) not null,
  unique(oracle_id, set_id)
);

/* Squire ID Model */

create table players (
  player_id uuid primary key, /* An internal id for the player */
  email varchar(255) not null,
  hashed_pwd varchar(255) not null,
  pwd_salt varchar(255) not null,

  /* Squire Core Settings */
  name varchar(255) not null unique,
  filtered_name varchar(255) not null,
  default_game_name varchar(255) not null default = '',

  /* Third Party Integrations (for identification from unregistered users + account linking) */
  mtga_username varchar(255) not null default = '',
  discord_id bigint not null default = -1 /* discord integration */
);

/* Squire Core Model */

create table tournament_settings (
  settings_id uuid primary key,
  make_vc boolean not null,
  make_tc boolean not null,
  trice_bot boolean not null,
  spectators_allowed boolean not null,
  spectators_can_see_hands boolean not null,
  spectators_can_chat boolean not null,
  only_registered boolean not null,

  format varchar(50) not null;
  match_duration integer check(tournament_settings.match_duration > 0) not null,
  match_players integer check(tournament_settings.match_players > 0) not null,

  /* No constraint as some people may have very funny match making */  
  win_points integer not null,
  lose_points integer not null,
  draw_points integer not null,
);

create table tournaments (
	tourn_id uuid primary key,
  settings_id uuid references tournament_settings(settings_id) not null,
  create_time timestamp not null,
  end_time timestamp,
  boolean in_progress not null, /* null in sql is often odd so I added this */
  boolean registration_open not null,
  varchar(255) name not null,
  varchar(255) filtered_name not null
);

create table decks (
	uuid deck_id primary key,
  uuid player_id references players(player_id) not null,
  varchar(255) name not null,
  varchar(255) filtered_name not null,
  varchar(8) deck_hash not null, /* for searching */
);

create table deck_cards (
	uuid deck_id references decks(deck_id) not null,
  uuid oracle_id not null
);

create table matches (
	match_id uuid primary key,
  tourn_id uuid references tournaments(tourn_id) not null,
  create_time timestamp not null default = CURRENT_TIMETAMP,
  in_progress boolean not null default = true, /* null in sql is often odd so I added this */
  finish_time timestamp,
  cancelled boolean not null default = false,
  match_number integer not null,
  unique(match_id, tourn_id)
);

create table match_players (
	match_id uuid references matches(match_id) not null,
  player_id uuid references players(player_id) not null,
  player_status integer not null default = 0, /* This is a software defined enum, 0 = lose, 1 = win, 2 = draw */
  unique(match_id, player_id)
);

/* Squire Bot Model */

create table guilds (
	guild_id bigint primary key,
  guild_name varchar(255) not null, /* A cache of the guild name */
  default_settings_id uuid references tournament_settings(settings_id) not null
);

create table discord_tournaments (
	toun_id uuid references tournaments(toun_id) not null unique,
  guild_id uuid references guilds(guild_id) not null,
  unique(tourn_id)
);

create table discord_players(				
	player_id uuid references players(player_id) not null,
  discord_id bigint not null,
  unique(player_id, discord_id)
);


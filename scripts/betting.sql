CREATE TABLE IF NOT EXISTS `schedule` (
  schedule_id        INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  home_team          VARCHAR(20),
  away_team          VARCHAR(20),
  home_team_win_odds FLOAT(4,3),
  away_team_win_odds FLOAT(4,3),
  tied_odds          FLOAT(4,3),
  schedule_time      DATETIME,
  schedule_group     VARCHAR(20),
  schedule_type      SMALLINT,
  schedule_status    SMALLINT,
  disable_betting    SMALLINT
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `bet` (
  user_id        varchar(200),
  schedule_id    int,
  betting_monery int,
  betting_result int,
  betting_odds   float(4,3),
  PRIMARY KEY (user_id, schedule_id)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;
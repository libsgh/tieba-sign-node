CREATE TABLE IF NOT EXISTS "tieba" (
  "guid" text NOT NULL,
  "uid" text,
  "uname" text,
  "fid" text,
  "fname" text,
  "level_id" text,
  "cur_score" text,
  "level_name" text,
  "levelup_score" text,
  "avatar" text,
  "slogan" text,
  "signTime" integer,
  "error_code" text,
  "ret_msg" text,
  PRIMARY KEY ("guid")
);
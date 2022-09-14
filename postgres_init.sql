CREATE TABLE IF NOT EXISTS tieba (
   guid varchar NOT NULL,
   uid varchar NULL,
   uname varchar NULL,
   fid varchar NULL,
   fname varchar NULL,
   level_id varchar NULL,
   cur_score varchar NULL,
   level_name varchar NULL,
   levelup_score varchar NULL,
   avatar varchar NULL,
   slogan varchar NULL,
   signtime _int4 NULL,
   error_code varchar NULL,
   ret_msg varchar NULL,
   CONSTRAINT tieba_pk PRIMARY KEY (guid)
);
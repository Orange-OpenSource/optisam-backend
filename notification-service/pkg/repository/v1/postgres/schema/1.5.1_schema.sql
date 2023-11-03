-- +migrate Up
CREATE TABLE dead_letter_queue (
 id serial PRIMARY KEY,
 topic VARCHAR NOT NULL,
 no_of_retries int NOT NULL,
 message jsonb NOT NULL,
 created_on TIMESTAMP NOT NULL DEFAULT NOW()
);


-- +migrate Down
drop table dead_letter_queue;

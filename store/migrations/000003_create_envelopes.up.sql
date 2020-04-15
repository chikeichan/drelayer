CREATE TABLE envelopes (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT                   NOT NULL REFERENCES users(id),
    guid       VARCHAR(16)              NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    refhash    VARCHAR                  NOT NULL DEFAULT '0000000000000000000000000000000000000000000000000000000000000000',
    UNIQUE (user_id, guid)
);
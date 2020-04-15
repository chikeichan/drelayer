CREATE TABLE posts (
    id          BIGSERIAL PRIMARY KEY,
    envelope_id BIGINT REFERENCES envelopes(id),
    body        TEXT       NOT NULL,
    title       VARCHAR,
    reference   VARCHAR,
    topic       VARCHAR,
    tags        varchar[]
);
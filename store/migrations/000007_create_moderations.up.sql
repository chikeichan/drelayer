CREATE TABLE moderations (
    id              BIGSERIAL PRIMARY KEY,
    envelope_id     BIGINT REFERENCES envelopes(id),
    reference       VARCHAR NOT NULL,
    moderation_type VARCHAR NOT NULL
);
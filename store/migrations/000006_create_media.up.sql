CREATE TABLE media (
    id          BIGSERIAL PRIMARY KEY,
    envelope_id BIGINT  NOT NULL REFERENCES envelopes(id),
    filename    VARCHAR NOT NULL,
    mime_type   VARCHAR NOT NULL,
    content     bytea   NOT NULL
);
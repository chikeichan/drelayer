CREATE TABLE connections (
    id              BIGSERIAL PRIMARY KEY,
    envelope_id     BIGINT REFERENCES envelopes(id),
    tld             VARCHAR NOT NULL,
    connection_type INT     NOT NULL,
    subdomain       VARCHAR
);
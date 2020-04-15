CREATE TABLE users (
    id                   BIGSERIAL PRIMARY KEY,
    tld_id               BIGINT                   NOT NULL REFERENCES tlds(id),
    username             VARCHAR(15)              NOT NULL,
    email                VARCHAR,
    hashed_password      VARCHAR                  NOT NULL,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    api_token_created_at TIMESTAMP WITH TIME ZONE,
    api_token            VARCHAR,
    UNIQUE (tld_id, username),
    UNIQUE (email),
    UNIQUE (api_token)
);
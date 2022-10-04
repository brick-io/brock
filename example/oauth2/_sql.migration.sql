CREATE TABLE IF NOT EXISTS public.users (
    _id     TEXT PRIMARY KEY,
    _name   TEXT UNIQUE NOT NULL,
);

CREATE TABLE IF NOT EXISTS public.clients (
    _id         TEXT PRIMARY KEY,
    _secret     TEXT NOT NULL,
    _domain     TEXT NOT NULL,
    _user_id    TEXT NOT NULL FOREIGN KEY REFERENCES public.users(_id),
);

CREATE TABLE IF NOT EXISTS public.tokens (
    _id             TEXT PRIMARY KEY,
    _client_id      TEXT NOT NULL FOREIGN KEY REFERENCES public.clients(_id),
    _user_id        TEXT NOT NULL FOREIGN KEY REFERENCES public.users(_id),
    _redirect_uri   TEXT NOT NULL,
    _scope          TEXT NOT NULL,
);

CREATE TABLE IF NOT EXISTS public.token_keys (
    _token_id   TEXT NOT NULL FOREIGN KEY REFERENCES public.tokens(_id),
    _type       INT NOT NULL,   -- either authorization-code (1) / access-token (2) / refresh-token (3)
    _value      TEXT UNIQUE NOT NULL,
    _created_at TIMESTAMPTZ NOT NULL,
    _expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (_token_id, _type)  -- limit to only 1 type per token
);
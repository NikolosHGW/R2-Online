-- R2 Online emulator — initial schema
-- PostgreSQL 16

BEGIN;

-- ─────────────────────────────────────────────────────────────────────────────
-- Accounts
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS accounts (
    id          SERIAL PRIMARY KEY,
    login       VARCHAR(64)  UNIQUE NOT NULL,
    -- Store bcrypt or argon2id hash, never plain text
    password    VARCHAR(256) NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    banned_at   TIMESTAMPTZ  NULL -- set when account is banned
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Game servers (the list shown in the lobby)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS game_servers (
    id          SERIAL       PRIMARY KEY,
    server_id   SMALLINT     NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    server_ip   VARCHAR(15)  NOT NULL,
    server_port INT          NOT NULL,
    type        SMALLINT     NOT NULL DEFAULT 1, -- 1=normal, 2=open
    hidden      BOOLEAN      NOT NULL DEFAULT FALSE,
    status      BOOLEAN      NOT NULL DEFAULT TRUE,  -- true=online
    congestion  SMALLINT     NOT NULL DEFAULT 0      -- 0-100
);

-- Seed a default server entry
INSERT INTO game_servers (server_id, name, server_ip, server_port, type, status)
VALUES (1, 'R2 Online', '127.0.0.1', 5000, 1, true)
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Sessions (Redis is the primary store; this is a durable audit log)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS sessions (
    id          BIGSERIAL    PRIMARY KEY,
    account_id  INT          NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    server_id   INT          NULL     REFERENCES game_servers(id),
    token       INT          NOT NULL, -- same token written to Redis
    in_game     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    expires_at  TIMESTAMPTZ  DEFAULT NOW() + INTERVAL '10 minutes'
);

CREATE INDEX IF NOT EXISTS idx_sessions_account ON sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token   ON sessions(token);

-- ─────────────────────────────────────────────────────────────────────────────
-- Characters
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS characters (
    id          SERIAL       PRIMARY KEY,
    owner_id    INT          NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    slot        SMALLINT     NOT NULL CHECK (slot BETWEEN 0 AND 2),
    nickname    VARCHAR(16)  NOT NULL,
    class       SMALLINT     NOT NULL,
    gender      SMALLINT     NOT NULL DEFAULT 0,
    head        SMALLINT     NOT NULL DEFAULT 0,
    face        SMALLINT     NOT NULL DEFAULT 0,
    body        SMALLINT     NOT NULL DEFAULT 0,
    home_map_no INT          NOT NULL DEFAULT 1,
    home_pos_x  FLOAT        NOT NULL DEFAULT 0,
    home_pos_y  FLOAT        NOT NULL DEFAULT 0,
    home_pos_z  FLOAT        NOT NULL DEFAULT 0,
    reg_date    TIMESTAMPTZ  DEFAULT NOW(),
    del_date    TIMESTAMPTZ  NULL, -- soft delete
    UNIQUE (owner_id, slot)
);

CREATE INDEX IF NOT EXISTS idx_characters_owner ON characters(owner_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- Character state (mutable, updated on logout)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS pc_state (
    character_id    INT      PRIMARY KEY REFERENCES characters(id) ON DELETE CASCADE,
    level           SMALLINT NOT NULL DEFAULT 1,
    exp             BIGINT   NOT NULL DEFAULT 0,
    hp              INT      NOT NULL DEFAULT 100,
    hp_add          INT      NOT NULL DEFAULT 0,
    mp              INT      NOT NULL DEFAULT 100,
    mp_add          INT      NOT NULL DEFAULT 0,
    strength        INT      NOT NULL DEFAULT 10,
    intelligence    INT      NOT NULL DEFAULT 10,
    dexterity       INT      NOT NULL DEFAULT 10,
    map_no          INT      NOT NULL DEFAULT 1,
    pos_x           FLOAT    NOT NULL DEFAULT 0,
    pos_y           FLOAT    NOT NULL DEFAULT 0,
    pos_z           FLOAT    NOT NULL DEFAULT 0,
    stomach         SMALLINT NOT NULL DEFAULT 1000,
    pk_count        INT      NOT NULL DEFAULT 0,
    chaotic         INT      NOT NULL DEFAULT 0,
    last_login      TIMESTAMPTZ NULL,
    last_logout     TIMESTAMPTZ NULL
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Inventory
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS pc_inventory (
    serial_no       BIGSERIAL    PRIMARY KEY,
    character_id    INT          NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_no         INT          NOT NULL,
    cnt             INT          NOT NULL DEFAULT 1,
    cnt_use         SMALLINT     NOT NULL DEFAULT 0,
    status          SMALLINT     NOT NULL DEFAULT 0,
    is_confirm      BOOLEAN      NOT NULL DEFAULT TRUE,
    end_date        TIMESTAMPTZ  NULL,
    binding_type    SMALLINT     NOT NULL DEFAULT 0,
    hole_count      SMALLINT     NOT NULL DEFAULT 0,
    apply_abn_item_no INT        NOT NULL DEFAULT 0,
    reg_date        TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_char ON pc_inventory(character_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- Equipped items
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS pc_equip (
    character_id    INT          NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    slot            SMALLINT     NOT NULL,
    serial_no       BIGINT       NOT NULL REFERENCES pc_inventory(serial_no) ON DELETE CASCADE,
    reg_date        TIMESTAMPTZ  DEFAULT NOW(),
    PRIMARY KEY (character_id, slot)
);

COMMIT;

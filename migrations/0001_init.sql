-- 0001_init: full v1 schema from docs/api.md
-- Projects carry a single nullable share_token (one active link per Project).
-- Pins store relative % coords; threads are single-level (parent_id -> root only).
-- Status machine (open/resolved/reopened) + 15-min client edit window enforced in app code.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE designers (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email       CITEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  name        TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- CITEXT needs the citext extension; fall back if unavailable:
-- If `CREATE EXTENSION citext` fails in your environment, change email to TEXT.

CREATE TABLE projects (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id          UUID NOT NULL REFERENCES designers(id) ON DELETE CASCADE,
  name              TEXT NOT NULL,
  description       TEXT NOT NULL DEFAULT '',
  share_token       TEXT UNIQUE,
  share_created_at  TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX projects_owner_idx ON projects(owner_id);

CREATE TABLE designs (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX designs_project_idx ON designs(project_id);

CREATE TABLE versions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  design_id   UUID NOT NULL REFERENCES designs(id) ON DELETE CASCADE,
  number      INTEGER NOT NULL,
  object_key  TEXT NOT NULL,
  width       INTEGER NOT NULL,
  height      INTEGER NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (design_id, number)
);
CREATE INDEX versions_design_idx ON versions(design_id);

CREATE TABLE pins (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  version_id  UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
  x           NUMERIC(5,2) NOT NULL CHECK (x >= 0 AND x <= 100),
  y           NUMERIC(5,2) NOT NULL CHECK (y >= 0 AND y <= 100),
  status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','resolved','reopened')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX pins_version_idx ON pins(version_id);
CREATE INDEX pins_status_idx ON pins(version_id, status);

CREATE TABLE comments (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  pin_id        UUID NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
  parent_id     UUID REFERENCES comments(id) ON DELETE CASCADE,
  body          TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 5000),
  author_type   TEXT NOT NULL CHECK (author_type IN ('client','designer')),
  designer_id   UUID REFERENCES designers(id) ON DELETE SET NULL,
  display_name  TEXT NOT NULL,
  client_token  TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX comments_pin_idx ON comments(pin_id);

-- internal/data/migrations/0001_init.sql
-- SQLite-совместимая схема (MVP). Постгрес-миграцию дадим отдельным файлом позже.

PRAGMA foreign_keys = ON;

-- ------------------------------------------------
-- Tenants (мульти-тенантность)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS tenants (
  id         TEXT PRIMARY KEY,                -- UUID/строка
  name       TEXT NOT NULL,
  timezone   TEXT,
  industry   TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- индексы
CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);

-- ------------------------------------------------
-- Clients (клиенты)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS clients (
  id            TEXT PRIMARY KEY,            -- UUID/строка
  tenant_id     TEXT NOT NULL,
  name          TEXT,
  phone         TEXT,
  loyalty_level TEXT,
  total_spent   REAL DEFAULT 0,
  first_seen    DATETIME,
  last_seen     DATETIME,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_clients_tenant ON clients(tenant_id);
CREATE INDEX IF NOT EXISTS idx_clients_phone  ON clients(phone);

-- ------------------------------------------------
-- Messages (входящие/исходящие)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS messages (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id  TEXT NOT NULL,
  client_id  TEXT,
  sender     TEXT NOT NULL CHECK (sender IN ('bot','user','operator')),
  channel    TEXT,                              -- 'wa','tg','ig','web','email'
  text       TEXT,
  intent     TEXT,
  sentiment  TEXT,
  lead_stage TEXT,                              -- 'cold','warm','hot' и т.п.
  ts         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_tenant_ts ON messages(tenant_id, ts);
CREATE INDEX IF NOT EXISTS idx_messages_client    ON messages(client_id);

-- ------------------------------------------------
-- Bookings (бронирования/сделки)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS bookings (
  id          TEXT PRIMARY KEY,                 -- UUID/строка
  tenant_id   TEXT NOT NULL,
  client_id   TEXT,
  start_time  DATETIME NOT NULL,
  end_time    DATETIME,
  status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','confirmed','cancelled','no_show')),
  amount      REAL NOT NULL DEFAULT 0,
  seats       INTEGER,
  source      TEXT,                             -- 'wa','tg','ig','web','pos'
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_bookings_tenant_time ON bookings(tenant_id, start_time);
CREATE INDEX IF NOT EXISTS idx_bookings_client      ON bookings(client_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status      ON bookings(status);

-- ------------------------------------------------
-- analytics_daily (агрегаты по дню)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS analytics_daily (
  tenant_id        TEXT NOT NULL,
  date             TEXT NOT NULL,               -- 'YYYY-MM-DD'
  total_clients    INTEGER DEFAULT 0,
  total_revenue    REAL    DEFAULT 0,
  avg_check        REAL    DEFAULT 0,
  bookings_count   INTEGER DEFAULT 0,
  conversion_rate  REAL    DEFAULT 0,           -- 0..1
  repeat_rate      REAL    DEFAULT 0,           -- 0..1
  peak_hour        INTEGER,                     -- 0..23
  weather_score    REAL,                        -- нормализованный показатель
  content_score    REAL,                        -- нормализованный показатель
  PRIMARY KEY (tenant_id, date),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- ------------------------------------------------
-- social_stats (интеграции соцсетей)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS social_stats (
  tenant_id      TEXT NOT NULL,
  date           TEXT NOT NULL,                 -- 'YYYY-MM-DD'
  platform       TEXT NOT NULL,                 -- 'instagram','tiktok','youtube'
  reach          INTEGER,
  impressions    INTEGER,
  clicks         INTEGER,
  views          INTEGER,
  ctr            REAL,
  content_count  INTEGER,
  PRIMARY KEY (tenant_id, date, platform),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- ------------------------------------------------
-- ai_recommendations (AI-советы владельцу)
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS ai_recommendations (
  id                   INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id            TEXT NOT NULL,
  date                 TEXT NOT NULL,           -- 'YYYY-MM-DD'
  topic                TEXT,                    -- 'sales','marketing','staff','pricing'...
  recommendation_text  TEXT NOT NULL,
  score                REAL,                    -- важность/уверенность
  meta_json            TEXT,                    -- произвольные детали в JSON
  created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ai_reco_tenant_date ON ai_recommendations(tenant_id, date);

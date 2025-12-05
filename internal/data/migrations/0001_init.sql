-- 0001_init.sql

-- Tenants (multi-tenant база)
CREATE TABLE IF NOT EXISTS tenants (
  tenant_id      TEXT PRIMARY KEY,
  owner_id       TEXT,
  business_name  TEXT NOT NULL,
  timezone       TEXT DEFAULT 'UTC',
  industry       TEXT,
  created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Клиенты
CREATE TABLE IF NOT EXISTS clients (
  client_id     TEXT PRIMARY KEY,
  tenant_id     TEXT NOT NULL,
  name          TEXT,
  phone         TEXT,
  loyalty_level TEXT,
  total_spent   REAL DEFAULT 0,
  first_seen    TIMESTAMP,
  last_seen     TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_clients_tenant ON clients(tenant_id);

-- Сообщения диалогов
CREATE TABLE IF NOT EXISTS messages (
  msg_id     TEXT PRIMARY KEY,
  tenant_id  TEXT NOT NULL,
  client_id  TEXT NOT NULL,
  sender     TEXT NOT NULL,            -- 'user' | 'bot' | 'operator'
  text       TEXT,
  channel    TEXT,                      -- 'wa' | 'tg' | ...
  ts         TIMESTAMP NOT NULL,
  intent     TEXT,
  sentiment  TEXT,
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id),
  FOREIGN KEY (client_id) REFERENCES clients(client_id)
);
CREATE INDEX IF NOT EXISTS idx_messages_tenant_ts ON messages(tenant_id, ts);
CREATE INDEX IF NOT EXISTS idx_messages_client_ts ON messages(client_id, ts);

-- Бронирования/заказы
CREATE TABLE IF NOT EXISTS bookings (
  booking_id  TEXT PRIMARY KEY,
  tenant_id   TEXT NOT NULL,
  client_id   TEXT NOT NULL,
  start_time  TIMESTAMP NOT NULL,
  end_time    TIMESTAMP,
  status      TEXT NOT NULL,            -- 'created' | 'confirmed' | 'cancelled' | 'no_show'
  amount      REAL DEFAULT 0,
  seats       INTEGER DEFAULT 1,
  source      TEXT,                      -- 'wa' | 'tg' | 'ig' | 'web'
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id),
  FOREIGN KEY (client_id) REFERENCES clients(client_id)
);
CREATE INDEX IF NOT EXISTS idx_bookings_tenant_time ON bookings(tenant_id, start_time);
CREATE INDEX IF NOT EXISTS idx_bookings_client ON bookings(client_id);

-- Суточная аналитика (агрегаты)
CREATE TABLE IF NOT EXISTS analytics_daily (
  tenant_id        TEXT NOT NULL,
  date             DATE NOT NULL,
  total_clients    INTEGER DEFAULT 0,
  bookings_count   INTEGER DEFAULT 0,
  total_revenue    REAL DEFAULT 0,
  avg_check        REAL DEFAULT 0,
  conversion_rate  REAL DEFAULT 0,      -- 0..1
  repeat_rate      REAL DEFAULT 0,      -- 0..1
  peak_hour        INTEGER,             -- 0..23
  weather_score    REAL,                -- нормализованный индекс
  content_score    REAL,                -- нормализованный индекс
  PRIMARY KEY (tenant_id, date),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id)
);

-- Соцметрики по дням
CREATE TABLE IF NOT EXISTS social_stats (
  tenant_id       TEXT NOT NULL,
  date            DATE NOT NULL,
  ig_reach        INTEGER,
  ig_impressions  INTEGER,
  ig_clicks       INTEGER,
  tiktok_views    INTEGER,
  tiktok_ctr      REAL,
  content_count   INTEGER,
  PRIMARY KEY (tenant_id, date),
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id)
);

-- AI-рекомендации по дням
CREATE TABLE IF NOT EXISTS ai_recommendations (
  id                 TEXT PRIMARY KEY,
  tenant_id          TEXT NOT NULL,
  date               DATE NOT NULL,
  topic              TEXT,
  recommendation_text TEXT NOT NULL,
  score              REAL,                -- важность/уверенность
  meta_json          TEXT,                -- произвольные детали (JSON)
  created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id)
);
CREATE INDEX IF NOT EXISTS idx_ai_reco_tenant_date ON ai_recommendations(tenant_id, date);

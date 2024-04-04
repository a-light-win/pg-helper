
CREATE TYPE DB_TASK_STATUS AS ENUM (
  'pending',
  'running',
  'completed',
  'failed',
  'cancelled'
);

CREATE TYPE DB_ACTION AS ENUM (
  'create',
  'drop',
  'backup',
  'remote-backup',
  'migrate',
  'restore'
);

CREATE TABLE IF NOT EXISTS db_tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  db_id BIGINT NOT NULL,
  db_name TEXT NOT NULL,
  action DB_ACTION NOT NULL,
  reason TEXT NOT NULL,
  status DB_TASK_STATUS NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
  updated_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
  data JSONB
);

CREATE INDEX db_tasks_db_id_idx ON db_tasks (db_id);
CREATE INDEX db_tasks_status_idx ON db_tasks (status);

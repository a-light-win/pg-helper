
CREATE TABLE IF NOT EXISTS db_task_depends (
  id BIGSERIAL PRIMARY KEY,
  task_id UUID NOT NULL,
  depends_on_task_id UUID NOT NULL
);

CREATE INDEX db_task_depends_task_id_idx ON db_task_depends (task_id);
CREATE INDEX db_task_depends_on_task_id_idx ON db_task_depends (depends_on_task_id);

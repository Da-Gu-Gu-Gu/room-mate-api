CREATE TABLE tasks (
    id uuid PRIMARY KEY,
    room_id uuid NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    name varchar(200) NOT NULL CHECK (length(trim(name)) > 0),
    description text,
    schedule_type text NOT NULL CHECK (schedule_type IN ('one_time', 'recurring')),
    created_by uuid NOT NULL REFERENCES users(id),
    first_due_at timestamptz NOT NULL,
    repeat_until timestamptz,
    frequency text CHECK (frequency IN ('daily', 'weekly')),
    timezone text NOT NULL,
    grace_period_minutes integer NOT NULL DEFAULT 1440 CHECK (grace_period_minutes BETWEEN 0 AND 10080),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'archived')),
    next_due_at timestamptz,
    last_order_no integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((schedule_type = 'one_time' AND frequency IS NULL AND repeat_until IS NULL)
        OR (schedule_type = 'recurring' AND frequency IS NOT NULL)),
    CHECK (repeat_until IS NULL OR repeat_until >= first_due_at)
);
CREATE INDEX tasks_room_id_idx ON tasks(room_id);
CREATE INDEX tasks_next_due_idx ON tasks(next_due_at) WHERE status = 'active';

CREATE TABLE task_members (
    id uuid PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    member_id uuid NOT NULL REFERENCES room_members(id),
    order_no integer NOT NULL CHECK (order_no > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(task_id, member_id),
    UNIQUE(task_id, order_no)
);

CREATE TABLE task_occurrences (
    id uuid PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    due_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL CHECK (expires_at >= due_at),
    assigned_member_id uuid NOT NULL REFERENCES room_members(id),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'missed', 'cancelled')),
    completed_by_member_id uuid REFERENCES room_members(id),
    completed_at timestamptz,
    proof_url text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(task_id, due_at),
    CHECK ((status = 'completed' AND completed_by_member_id IS NOT NULL AND completed_at IS NOT NULL)
        OR (status <> 'completed' AND completed_by_member_id IS NULL AND completed_at IS NULL AND proof_url IS NULL))
);
CREATE INDEX task_occurrences_pending_expiry_idx ON task_occurrences(expires_at) WHERE status = 'pending';
CREATE INDEX task_occurrences_assignee_idx ON task_occurrences(assigned_member_id, due_at);

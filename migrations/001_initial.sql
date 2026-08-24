CREATE TABLE IF NOT EXISTS release_operators (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    role text NOT NULL CHECK (role IN ('managed_device_operator', 'installation_report_release_operator', 'quality_reviewer', 'safety_supervisor')),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rollout_campaigns (
    id uuid PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL,
    status text NOT NULL CHECK (status IN ('draft', 'scheduled', 'active', 'closed')),
    timezone text NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_by uuid NOT NULL REFERENCES release_operators(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS managed_devices (
    id uuid PRIMARY KEY,
    rollout_campaign_id uuid NOT NULL REFERENCES rollout_campaigns(id) ON DELETE CASCADE,
    code text NOT NULL,
    rollout_lane text NOT NULL,
    required_successes integer NOT NULL CHECK (required_successes > 0),
    completed_installs integer NOT NULL DEFAULT 0 CHECK (completed_installs >= 0),
    UNIQUE (rollout_campaign_id, code)
);

CREATE TABLE IF NOT EXISTS assignments (
    id uuid PRIMARY KEY,
    rollout_campaign_id uuid NOT NULL REFERENCES rollout_campaigns(id) ON DELETE CASCADE,
    managed_device_id uuid NOT NULL REFERENCES managed_devices(id) ON DELETE CASCADE,
    release_operator_id uuid NOT NULL REFERENCES release_operators(id),
    status text NOT NULL CHECK (status IN ('queued','active','completed','cancelled')),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    UNIQUE (managed_device_id, starts_at),
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS deployment_jobs (
    id uuid PRIMARY KEY,
    rollout_campaign_id uuid NOT NULL REFERENCES rollout_campaigns(id),
    managed_device_id uuid NOT NULL REFERENCES managed_devices(id),
    task_code text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('queued', 'completed', 'activation_pending', 'accepted', 'in_progress', 'verified', 'rejected', 'archived')),
    completed_at timestamptz,
    accepted_at timestamptz,
    expires_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS activation_events (
    id uuid PRIMARY KEY,
    deployment_job_id uuid NOT NULL REFERENCES deployment_jobs(id) ON DELETE CASCADE,
    from_operator uuid REFERENCES release_operators(id),
    to_operator uuid NOT NULL REFERENCES release_operators(id),
    location text NOT NULL,
    recorded_at timestamptz NOT NULL,
    note text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS rollout_waves (
    id uuid PRIMARY KEY,
    code text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'cancelled')),
    method text NOT NULL,
    capacity integer NOT NULL CHECK (capacity > 0),
    started_at timestamptz,
    completed_at timestamptz,
    version bigint NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS rollout_wave_items (
    rollout_wave_id uuid NOT NULL REFERENCES rollout_waves(id) ON DELETE CASCADE,
    deployment_job_id uuid NOT NULL REFERENCES deployment_jobs(id),
    PRIMARY KEY (rollout_wave_id, deployment_job_id)
);

CREATE TABLE IF NOT EXISTS installation_reports (
    id uuid PRIMARY KEY,
    deployment_job_id uuid NOT NULL REFERENCES deployment_jobs(id),
    rollout_wave_id uuid NOT NULL REFERENCES rollout_waves(id),
    recorded_by uuid NOT NULL REFERENCES release_operators(id),
    status text NOT NULL CHECK (status IN ('pending', 'verified', 'rejected')),
    value numeric(18,6) NOT NULL,
    unit text NOT NULL,
    limit_value numeric(18,6) NOT NULL,
    measured_at timestamptz NOT NULL,
    reviewed_at timestamptz,
    version bigint NOT NULL DEFAULT 1,
    UNIQUE (deployment_job_id, rollout_wave_id)
);

CREATE TABLE IF NOT EXISTS health_alerts (
    id uuid PRIMARY KEY,
    deployment_job_id uuid NOT NULL REFERENCES deployment_jobs(id),
    kind text NOT NULL CHECK (kind IN ('retask', 'repeat_managed_device', 'safety_adjustment', 'close_record')),
    status text NOT NULL CHECK (status IN ('open', 'in_progress', 'closed')),
    reason text NOT NULL,
    due_at timestamptz NOT NULL,
    closed_at timestamptz,
    UNIQUE (deployment_job_id, kind, status)
);

CREATE TABLE IF NOT EXISTS audit_events (
    id uuid PRIMARY KEY,
    request_id text NOT NULL,
    release_operator_id uuid REFERENCES release_operators(id),
    object_type text NOT NULL,
    object_id uuid NOT NULL,
    action text NOT NULL,
    outcome text NOT NULL,
    detail jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key text PRIMARY KEY,
    request_hash text NOT NULL,
    response_code integer NOT NULL,
    response_body jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS deployment_jobs_rollout_campaign_status_idx ON deployment_jobs(rollout_campaign_id, status);
CREATE INDEX IF NOT EXISTS deployment_jobs_expiry_idx ON deployment_jobs(status, expires_at);
CREATE INDEX IF NOT EXISTS activation_task_time_idx ON activation_events(deployment_job_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS installation_reports_status_idx ON installation_reports(status, measured_at);
CREATE INDEX IF NOT EXISTS health_alerts_due_idx ON health_alerts(status, due_at);
CREATE INDEX IF NOT EXISTS audit_object_idx ON audit_events(object_type, object_id, created_at DESC);

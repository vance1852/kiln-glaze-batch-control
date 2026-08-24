CREATE TABLE console_users (
    id uuid PRIMARY KEY,
    username text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    real_name text NOT NULL,
    phone text NOT NULL DEFAULT '',
    role text NOT NULL CHECK (role IN ('curator', 'release_operator', 'reviewer')),
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE console_sessions (
    token_hash text PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES console_users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);

CREATE TABLE console_managed_devices (
    id uuid PRIMARY KEY,
    title text NOT NULL,
    device_class integer NOT NULL CHECK (device_class IN (1, 2)),
    comrollout_campaigned_on date,
    accession_number text NOT NULL DEFAULT '',
    repository_code text NOT NULL DEFAULT '',
    storage_zone text NOT NULL DEFAULT '',
    donor_name text NOT NULL DEFAULT '',
    curator_contact text NOT NULL DEFAULT '',
    condition_status integer NOT NULL DEFAULT 1 CHECK (condition_status BETWEEN 1 AND 3),
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE console_release_operators (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    specialty_level integer NOT NULL CHECK (specialty_level IN (1, 2)),
    phone text NOT NULL DEFAULT '',
    skills text NOT NULL DEFAULT '',
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE console_rollout_profiles (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    risk_budget numeric(10,2) NOT NULL DEFAULT 0 CHECK (risk_budget >= 0),
    duration_minutes integer NOT NULL DEFAULT 60 CHECK (duration_minutes BETWEEN 1 AND 1440),
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE console_command_orders (
    id uuid PRIMARY KEY,
    work_order_no text NOT NULL UNIQUE,
    managed_device_id uuid NOT NULL REFERENCES console_managed_devices(id) ON DELETE RESTRICT,
    rollout_profile_id uuid NOT NULL REFERENCES console_rollout_profiles(id) ON DELETE RESTRICT,
    release_operator_id uuid REFERENCES console_release_operators(id) ON DELETE SET NULL,
    scheduled_at timestamptz,
    status integer NOT NULL DEFAULT 0 CHECK (status BETWEEN 0 AND 3),
    remark text NOT NULL DEFAULT '',
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE console_installation_reports (
    id uuid PRIMARY KEY,
    managed_device_id uuid NOT NULL REFERENCES console_managed_devices(id) ON DELETE RESTRICT,
    relative_humidity numeric(5,1),
    temperature_c numeric(5,1),
    illuminance_lux numeric(8,1),
    acidity_ph numeric(8,3),
    pest_index numeric(8,3),
    remark text NOT NULL DEFAULT '',
    recorded_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE console_logs (
    id uuid PRIMARY KEY,
    username text NOT NULL,
    operation text NOT NULL,
    method text NOT NULL,
    ip text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX console_sessions_user_expiry_idx ON console_sessions(user_id, expires_at) WHERE revoked_at IS NULL;
CREATE INDEX console_managed_devices_search_idx ON console_managed_devices(title, accession_number) WHERE deleted_at IS NULL;
CREATE INDEX console_release_operators_search_idx ON console_release_operators(name, phone) WHERE deleted_at IS NULL;
CREATE INDEX console_rollout_profiles_search_idx ON console_rollout_profiles(name) WHERE deleted_at IS NULL;
CREATE INDEX console_command_orders_status_time_idx ON console_command_orders(status, scheduled_at DESC);
CREATE INDEX console_installation_reports_managed_device_time_idx ON console_installation_reports(managed_device_id, recorded_at DESC);
CREATE INDEX console_logs_time_idx ON console_logs(created_at DESC);

INSERT INTO console_users(id, username, password_hash, real_name, phone, role) VALUES
('10000000-0000-0000-0000-000000000001', 'admin', '$2a$10$.yXjTzrlVGyHtUlJ3DtbXu/BsE8n0891/6TRfIRXfwaT5b8fqT6Iu', 'Ground Control Admin', '13800138000', 'curator'),
('10000000-0000-0000-0000-000000000002', 'repairer', '$2a$10$cLYkgpV5AiyvLDG4AgWHYu7lmGAS7J8vbDB1QMYRBbB4yayIyhZPy', 'Installation Operator', '13800138001', 'release_operator'),
('10000000-0000-0000-0000-000000000003', 'reviewer', '$2a$10$1mw/cjufLNttJ2KZ4U/jGeAozEZWGz1XzqWLRjNvvrSqeS/1V6iG2', 'RolloutCampaign Reviewer', '13800138002', 'reviewer');

INSERT INTO console_managed_devices(id, title, device_class, comrollout_campaigned_on, accession_number, repository_code, storage_zone, donor_name, curator_contact, condition_status) VALUES
('20000000-0000-0000-0000-000000000001', 'Aurora-1 Earth Observer', 1, '2024-03-15', 'SAT-0001', 'GS-01', 'North antenna bay', 'Orbital Research Agency', '+86-10-10000001', 1),
('20000000-0000-0000-0000-000000000002', 'Beacon-7 Relay ManagedDevice', 2, '2024-07-22', 'SAT-0002', 'GS-02', 'Relay rack B-02', 'National Space Lab', '+86-10-10000002', 2),
('20000000-0000-0000-0000-000000000003', 'Lumen-3 Weather ManagedDevice', 1, '2025-01-08', 'SAT-0003', 'GS-03', 'North antenna bay', 'Orbital Research Agency', '+86-10-10000003', 1);

INSERT INTO console_release_operators(id, name, specialty_level, phone, skills) VALUES
('30000000-0000-0000-0000-000000000001', 'Wang Rui', 2, '13800138011', 'installation decoding, link scheduling'),
('30000000-0000-0000-0000-000000000002', 'Liu Qing', 2, '13800138012', 'command approval, key rotation'),
('30000000-0000-0000-0000-000000000003', 'Chen Mo', 1, '13800138013', 'health checks, anomaly triage');

INSERT INTO console_rollout_profiles(id, name, description, risk_budget, duration_minutes) VALUES
('40000000-0000-0000-0000-000000000001', 'Installation Calibration', 'Read and validate the current sensor packet', 20.00, 60),
('40000000-0000-0000-0000-000000000002', 'Attitude Correction', 'Apply a bounded attitude correction command', 45.00, 120),
('40000000-0000-0000-0000-000000000003', 'Relay Reconfiguration', 'Switch the redundant communications path', 65.00, 180),
('40000000-0000-0000-0000-000000000004', 'Safe Mode Exit', 'Release a managed_device from safe mode after review', 80.00, 240);

INSERT INTO console_command_orders(id, work_order_no, managed_device_id, rollout_profile_id, release_operator_id, scheduled_at, status, remark) VALUES
('50000000-0000-0000-0000-000000000001', 'CMD-20260819001', '20000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', now() + interval '1 hour', 0, 'Awaiting operator confirmation'),
('50000000-0000-0000-0000-000000000002', 'CMD-20260819002', '20000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000002', now() + interval '2 hours', 1, 'Approval in progress');

INSERT INTO console_installation_reports(id, managed_device_id, relative_humidity, temperature_c, illuminance_lux, acidity_ph, pest_index, remark, recorded_at) VALUES
('60000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 48, 20.5, 45, 6.4, 0.1, 'Nominal bus installation', now() - interval '2 hours'),
('60000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 61, 23.2, 70, 5.2, 0.4, 'Elevated relay temperature', now() - interval '1 hour');

INSERT INTO console_logs(id, username, operation, method, ip)
VALUES ('70000000-0000-0000-0000-000000000001', 'admin', 'initial seed', 'migration.003', '127.0.0.1');

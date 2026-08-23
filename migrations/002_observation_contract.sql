DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'installation_reports' AND column_name = 'value'
    ) THEN
        ALTER TABLE installation_reports RENAME COLUMN value TO risk_score;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'installation_reports' AND column_name = 'unit'
    ) THEN
        ALTER TABLE installation_reports RENAME COLUMN unit TO scale;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'installation_reports' AND column_name = 'limit_value'
    ) THEN
        ALTER TABLE installation_reports RENAME COLUMN limit_value TO alert_threshold;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'installation_reports' AND column_name = 'measured_at'
    ) THEN
        ALTER TABLE installation_reports RENAME COLUMN measured_at TO observed_at;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'installation_reports_risk_score_nonnegative') THEN
        ALTER TABLE installation_reports
            ADD CONSTRAINT installation_reports_risk_score_nonnegative CHECK (risk_score >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'installation_reports_alert_threshold_nonnegative') THEN
        ALTER TABLE installation_reports
            ADD CONSTRAINT installation_reports_alert_threshold_nonnegative CHECK (alert_threshold >= 0);
    END IF;
END $$;

ALTER TABLE health_alerts DROP CONSTRAINT IF EXISTS health_alerts_kind_check;
UPDATE health_alerts SET kind = 'reassess' WHERE kind = 'retask';
ALTER TABLE health_alerts
    ADD CONSTRAINT health_alerts_kind_check CHECK (kind IN ('reassess', 'repeat_managed_device', 'safety_adjustment', 'close_record'));

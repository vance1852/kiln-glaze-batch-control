package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type RolloutWaveJob struct {
	ID          string
	Attempts    int
	MaxAttempts int
}
type RolloutWaveExecutor interface {
	Execute(context.Context, RolloutWaveJob) error
}

type RolloutWaveProcessor struct {
	executor RolloutWaveExecutor
	policy   RetryPolicy
	logger   *slog.Logger
	metrics  *Metrics
}

func NewRolloutWaveProcessor(executor RolloutWaveExecutor, policy RetryPolicy, logger *slog.Logger, metrics *Metrics) *RolloutWaveProcessor {
	if logger == nil {
		logger = slog.Default()
	}
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &RolloutWaveProcessor{executor: executor, policy: policy, logger: logger, metrics: metrics}
}

func (p *RolloutWaveProcessor) Process(ctx context.Context, job RolloutWaveJob) error {
	if job.ID == "" {
		return fmt.Errorf("rollout_wave job id is required")
	}
	if p.executor == nil {
		return fmt.Errorf("rollout_wave executor is nil")
	}
	policy := p.policy
	if job.MaxAttempts > 0 && policy.Attempts > job.MaxAttempts {
		policy.Attempts = job.MaxAttempts
	}
	start := time.Now()
	err := RunWithRetry(ctx, policy, func(callCtx context.Context) error { job.Attempts++; return p.executor.Execute(callCtx, job) })
	p.metrics.RecordRun()
	if err != nil {
		p.metrics.RecordFailure()
		p.logger.Error("rollout_wave job failed", "rollout_wave_id", job.ID, "attempts", job.Attempts, "duration", time.Since(start), "error", err)
		return err
	}
	p.logger.Info("rollout_wave job completed", "rollout_wave_id", job.ID, "attempts", job.Attempts, "duration", time.Since(start))
	return nil
}

package console

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Store) InstallationReportPage(ctx context.Context, current, size int, managed_deviceID string) (Page[InstallationReport], error) {
	current, size, offset := pageBounds(current, size)
	page := Page[InstallationReport]{Records: make([]InstallationReport, 0), Current: current, Size: size}
	where := ""
	args := []any{size, offset}
	if managed_deviceID != "" {
		where = "WHERE h.managed_device_id=$3"
		args = append(args, managed_deviceID)
	}
	rows, err := s.pool.Query(ctx, `SELECT h.id,h.managed_device_id,e.title,h.relative_humidity,h.temperature_c,h.illuminance_lux,h.acidity_ph,h.pest_index,h.remark,h.recorded_at
		FROM console_installation_reports h JOIN console_managed_devices e ON e.id=h.managed_device_id `+where+` ORDER BY h.recorded_at DESC LIMIT $1 OFFSET $2`, args...)
	if err != nil {
		return page, fmt.Errorf("查询舱室环境记录: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item InstallationReport
		if err := rows.Scan(&item.ID, &item.ManagedDeviceID, &item.ManagedDeviceTitle, &item.RelativeHumidity, &item.TemperatureC, &item.IlluminanceLux, &item.AcidityPH, &item.PestIndex, &item.Remark, &item.RecordedAt); err != nil {
			return page, fmt.Errorf("读取舱室环境记录: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	countQuery := `SELECT count(*) FROM console_installation_reports`
	countArgs := []any{}
	if managed_deviceID != "" {
		countQuery += ` WHERE managed_device_id=$1`
		countArgs = append(countArgs, managed_deviceID)
	}
	if err := s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计舱室环境记录: %w", err)
	}
	return page, rows.Err()
}

func (s *Store) CreateInstallationReport(ctx context.Context, item InstallationReport) (InstallationReport, error) {
	item.ID = uuid.NewString()
	if item.RecordedAt.IsZero() {
		if err := s.pool.QueryRow(ctx, `INSERT INTO console_installation_reports(id,managed_device_id,relative_humidity,temperature_c,illuminance_lux,acidity_ph,pest_index,remark)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING recorded_at`, item.ID, item.ManagedDeviceID, item.RelativeHumidity, item.TemperatureC, item.IlluminanceLux, item.AcidityPH, item.PestIndex, item.Remark).Scan(&item.RecordedAt); err != nil {
			return InstallationReport{}, wrap("新增舱室环境记录", err)
		}
	} else {
		_, err := s.pool.Exec(ctx, `INSERT INTO console_installation_reports(id,managed_device_id,relative_humidity,temperature_c,illuminance_lux,acidity_ph,pest_index,remark,recorded_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, item.ID, item.ManagedDeviceID, item.RelativeHumidity, item.TemperatureC, item.IlluminanceLux, item.AcidityPH, item.PestIndex, item.Remark, item.RecordedAt)
		if err != nil {
			return InstallationReport{}, wrap("新增舱室环境记录", err)
		}
	}
	return item, nil
}

func (s *Store) LogPage(ctx context.Context, current, size int) (Page[Log], error) {
	current, size, offset := pageBounds(current, size)
	page := Page[Log]{Records: make([]Log, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT id,username,operation,method,ip,created_at FROM console_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询日志: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Log
		if err := rows.Scan(&item.ID, &item.Username, &item.Operation, &item.Method, &item.IP, &item.CreateTime); err != nil {
			return page, fmt.Errorf("读取日志: %w", err)
		}
		page.Records = append(page.Records, item)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_logs`).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计日志: %w", err)
	}
	return page, rows.Err()
}

package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) ManagedDevicePage(ctx context.Context, current, size int, keyword string) (Page[ManagedDevice], error) {
	current, size, offset := pageBounds(current, size)
	keyword = "%" + strings.TrimSpace(keyword) + "%"
	page := Page[ManagedDevice]{Records: make([]ManagedDevice, 0), Current: current, Size: size}
	rows, err := s.pool.Query(ctx, `SELECT id,title,device_class,comrollout_campaigned_on,accession_number,repository_code,storage_zone,donor_name,curator_contact,condition_status,status
		FROM console_managed_devices WHERE deleted_at IS NULL AND (title ILIKE $1 OR accession_number ILIKE $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, keyword, size, offset)
	if err != nil {
		return page, fmt.Errorf("查询馆藏芯片批次列表: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item ManagedDevice
		var date *time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.DeviceClass, &date, &item.AccessionNumber, &item.RepositoryCode, &item.StorageZone, &item.DonorName, &item.CuratorContact, &item.ConditionStatus, &item.Status); err != nil {
			return page, fmt.Errorf("读取馆藏芯片批次信息: %w", err)
		}
		item.Comrollout_campaignedOn = formatDate(date)
		page.Records = append(page.Records, item)
	}
	if err := rows.Err(); err != nil {
		return page, fmt.Errorf("读取馆藏芯片批次列表: %w", err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM console_managed_devices WHERE deleted_at IS NULL AND (title ILIKE $1 OR accession_number ILIKE $1)`, keyword).Scan(&page.Total); err != nil {
		return page, fmt.Errorf("统计馆藏芯片批次数量: %w", err)
	}
	return page, nil
}

func (s *Store) ManagedDeviceList(ctx context.Context) ([]ManagedDevice, error) {
	page, err := s.ManagedDevicePage(ctx, 1, 100, "")
	return page.Records, err
}

func (s *Store) ManagedDeviceByID(ctx context.Context, id string) (ManagedDevice, error) {
	var item ManagedDevice
	var date *time.Time
	err := s.pool.QueryRow(ctx, `SELECT id,title,device_class,comrollout_campaigned_on,accession_number,repository_code,storage_zone,donor_name,curator_contact,condition_status,status
		FROM console_managed_devices WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&item.ID, &item.Title, &item.DeviceClass, &date, &item.AccessionNumber, &item.RepositoryCode, &item.StorageZone, &item.DonorName, &item.CuratorContact, &item.ConditionStatus, &item.Status,
	)
	item.Comrollout_campaignedOn = formatDate(date)
	return item, wrap("查询馆藏芯片批次", err)
}

func (s *Store) CreateManagedDevice(ctx context.Context, item ManagedDevice) (ManagedDevice, error) {
	item.ID = uuid.NewString()
	_, err := s.pool.Exec(ctx, `INSERT INTO console_managed_devices(id,title,device_class,comrollout_campaigned_on,accession_number,repository_code,storage_zone,donor_name,curator_contact,condition_status,status)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, item.ID, strings.TrimSpace(item.Title), item.DeviceClass, parseDate(item.Comrollout_campaignedOn), item.AccessionNumber, item.RepositoryCode, item.StorageZone, item.DonorName, item.CuratorContact, item.ConditionStatus, item.Status)
	return item, wrap("新增馆藏芯片批次", err)
}

func (s *Store) UpdateManagedDevice(ctx context.Context, item ManagedDevice) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_managed_devices SET title=$1,device_class=$2,comrollout_campaigned_on=$3,accession_number=$4,repository_code=$5,storage_zone=$6,donor_name=$7,curator_contact=$8,condition_status=$9,status=$10,updated_at=now()
		WHERE id=$11 AND deleted_at IS NULL`, strings.TrimSpace(item.Title), item.DeviceClass, parseDate(item.Comrollout_campaignedOn), item.AccessionNumber, item.RepositoryCode, item.StorageZone, item.DonorName, item.CuratorContact, item.ConditionStatus, item.Status, item.ID)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("更新馆藏芯片批次: 记录不存在")
	}
	return wrap("更新馆藏芯片批次", err)
}

func (s *Store) DeleteManagedDevice(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE console_managed_devices SET deleted_at=now(),updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err == nil && tag.RowsAffected() != 1 {
		return fmt.Errorf("删除馆藏芯片批次: 记录不存在")
	}
	return wrap("删除馆藏芯片批次", err)
}

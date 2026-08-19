package pet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Service) AddRecord(ctx context.Context, p Principal, input DailyRecord) (DailyRecord, error) {
	if p.UserID == 0 {
		return DailyRecord{}, ErrUnauthenticated
	}
	order, err := s.GetOrder(ctx, p, input.OrderID)
	if err != nil {
		return DailyRecord{}, err
	}
	if p.Role != RoleAdmin && order.UserID != p.UserID {
		return DailyRecord{}, ErrForbidden
	}
	if order.Status != "IN_PROGRESS" {
		return DailyRecord{}, fmt.Errorf("%w: records require an in-progress order", ErrInvalidState)
	}
	if input.RecordDate.IsZero() {
		input.RecordDate = s.now()
	}
	now := formatStoredTime(s.now())
	result, err := s.store.db.ExecContext(ctx, `INSERT INTO daily_records(order_id,record_date,diet,defecation,activity,spirit,remarks,media_urls,create_time,update_time) VALUES(?,?,?,?,?,?,?,?,?,?)`, input.OrderID, input.RecordDate.UTC().Format("2006-01-02"), input.Diet, input.Defecation, input.Activity, input.Spirit, input.Remarks, input.MediaURLs, now, now)
	if err != nil {
		return DailyRecord{}, translateServiceError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return DailyRecord{}, err
	}
	return s.GetRecord(ctx, p, id)
}
func (s *Service) UpdateRecord(ctx context.Context, p Principal, input DailyRecord) (DailyRecord, error) {
	if input.ID == 0 {
		return DailyRecord{}, fmt.Errorf("%w: record id is required", ErrValidation)
	}
	existing, err := s.GetRecord(ctx, p, input.ID)
	if err != nil {
		return DailyRecord{}, err
	}
	if p.Role != RoleAdmin {
		order, err := s.GetOrder(ctx, p, existing.OrderID)
		if err != nil {
			return DailyRecord{}, err
		}
		if order.UserID != p.UserID {
			return DailyRecord{}, ErrForbidden
		}
	}
	if input.RecordDate.IsZero() {
		input.RecordDate = existing.RecordDate
	}
	_, err = s.store.db.ExecContext(ctx, `UPDATE daily_records SET record_date=?,diet=?,defecation=?,activity=?,spirit=?,remarks=?,media_urls=?,update_time=? WHERE record_id=?`, input.RecordDate.UTC().Format("2006-01-02"), input.Diet, input.Defecation, input.Activity, input.Spirit, input.Remarks, input.MediaURLs, formatStoredTime(s.now()), input.ID)
	if err != nil {
		return DailyRecord{}, translateServiceError(err)
	}
	return s.GetRecord(ctx, p, input.ID)
}
func (s *Service) DeleteRecord(ctx context.Context, p Principal, id int64) error {
	record, err := s.GetRecord(ctx, p, id)
	if err != nil {
		return err
	}
	if p.Role != RoleAdmin {
		order, err := s.GetOrder(ctx, p, record.OrderID)
		if err != nil {
			return err
		}
		if order.UserID != p.UserID {
			return ErrForbidden
		}
	}
	result, err := s.store.db.ExecContext(ctx, `DELETE FROM daily_records WHERE record_id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result, ErrNotFound)
}
func (s *Service) GetRecord(ctx context.Context, p Principal, id int64) (DailyRecord, error) {
	if p.UserID == 0 {
		return DailyRecord{}, ErrUnauthenticated
	}
	var item DailyRecord
	var date, created, updated string
	var ownerID int64
	err := s.store.db.QueryRowContext(ctx, `SELECT d.record_id,d.order_id,d.record_date,COALESCE(d.diet,''),COALESCE(d.defecation,''),COALESCE(d.activity,''),COALESCE(d.spirit,''),COALESCE(d.remarks,''),COALESCE(d.media_urls,''),d.create_time,d.update_time,o.user_id FROM daily_records d JOIN foster_orders o ON o.order_id=d.order_id WHERE d.record_id=?`, id).Scan(&item.ID, &item.OrderID, &date, &item.Diet, &item.Defecation, &item.Activity, &item.Spirit, &item.Remarks, &item.MediaURLs, &created, &updated, &ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		return DailyRecord{}, ErrNotFound
	} else if err != nil {
		return DailyRecord{}, err
	}
	if p.Role != RoleAdmin && ownerID != p.UserID {
		return DailyRecord{}, ErrForbidden
	}
	item.RecordDate, _ = timeParseDate(date)
	item.CreatedAt, _ = parseStoredTime(created)
	item.UpdatedAt, _ = parseStoredTime(updated)
	return item, nil
}
func (s *Service) RecordsByOrder(ctx context.Context, p Principal, orderID int64) ([]DailyRecord, error) {
	page, err := s.ListRecords(ctx, p, 1, 100, orderID, 0, "", "")
	return page.List, err
}
func (s *Service) ListRecords(ctx context.Context, p Principal, pageNum, pageSize int, orderID, userID int64, startDate, endDate string) (Page[DailyRecord], error) {
	if p.UserID == 0 {
		return Page[DailyRecord]{}, ErrUnauthenticated
	}
	pageNum, pageSize = normalizePage(pageNum, pageSize)
	where := []string{"1=1"}
	args := []any{}
	if p.Role != RoleAdmin {
		where = append(where, "o.user_id=?")
		args = append(args, p.UserID)
	} else if userID != 0 {
		where = append(where, "o.user_id=?")
		args = append(args, userID)
	}
	if orderID != 0 {
		where = append(where, "d.order_id=?")
		args = append(args, orderID)
	}
	if startDate != "" {
		where = append(where, "d.record_date>=?")
		args = append(args, startDate)
	}
	if endDate != "" {
		where = append(where, "d.record_date<=?")
		args = append(args, endDate)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := s.store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM daily_records d JOIN foster_orders o ON o.order_id=d.order_id WHERE "+clause, args...).Scan(&total); err != nil {
		return Page[DailyRecord]{}, err
	}
	query := "SELECT d.record_id,d.order_id,d.record_date,COALESCE(d.diet,''),COALESCE(d.defecation,''),COALESCE(d.activity,''),COALESCE(d.spirit,''),COALESCE(d.remarks,''),COALESCE(d.media_urls,''),d.create_time,d.update_time FROM daily_records d JOIN foster_orders o ON o.order_id=d.order_id WHERE " + clause + " ORDER BY d.record_date DESC,d.record_id DESC LIMIT ? OFFSET ?"
	rows, err := s.store.db.QueryContext(ctx, query, append(args, pageSize, (pageNum-1)*pageSize)...)
	if err != nil {
		return Page[DailyRecord]{}, err
	}
	defer rows.Close()
	items := make([]DailyRecord, 0)
	for rows.Next() {
		var item DailyRecord
		var date, created, updated string
		if err := rows.Scan(&item.ID, &item.OrderID, &date, &item.Diet, &item.Defecation, &item.Activity, &item.Spirit, &item.Remarks, &item.MediaURLs, &created, &updated); err != nil {
			return Page[DailyRecord]{}, err
		}
		item.RecordDate, _ = timeParseDate(date)
		item.CreatedAt, _ = parseStoredTime(created)
		item.UpdatedAt, _ = parseStoredTime(updated)
		items = append(items, item)
	}
	return Page[DailyRecord]{List: items, Total: total, PageNum: pageNum, PageSize: pageSize}, rows.Err()
}
func timeParseDate(value string) (time.Time, error) { return time.Parse("2006-01-02", value) }

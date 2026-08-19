package pet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type CreateOrderInput struct {
	PetID     int64          `json:"petId"`
	RoomID    int64          `json:"roomId"`
	StartTime time.Time      `json:"startTime"`
	EndTime   time.Time      `json:"endTime"`
	Remarks   string         `json:"remarks"`
	Services  []OrderService `json:"services"`
}

func (s *Service) CreateOrder(ctx context.Context, principal Principal, input CreateOrderInput) (FosterOrder, error) {
	if principal.UserID == 0 {
		return FosterOrder{}, ErrUnauthenticated
	}
	if input.PetID == 0 || input.RoomID == 0 || !input.EndTime.After(input.StartTime) {
		return FosterOrder{}, fmt.Errorf("%w: pet, room and valid date range are required", ErrValidation)
	}
	tx, err := s.store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return FosterOrder{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var ownerID int64
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM pets WHERE pet_id=?`, input.PetID).Scan(&ownerID); errors.Is(err, sql.ErrNoRows) {
		return FosterOrder{}, ErrNotFound
	} else if err != nil {
		return FosterOrder{}, err
	}
	if principal.Role != RoleAdmin && ownerID != principal.UserID {
		return FosterOrder{}, ErrForbidden
	}
	var roomType, status string
	var price float64
	var capacity int
	if err := tx.QueryRowContext(ctx, `SELECT room_type,status,price_per_day,capacity FROM rooms WHERE room_id=?`, input.RoomID).Scan(&roomType, &status, &price, &capacity); errors.Is(err, sql.ErrNoRows) {
		return FosterOrder{}, ErrNotFound
	} else if err != nil {
		return FosterOrder{}, err
	}
	if status == "CLEANING" {
		return FosterOrder{}, fmt.Errorf("%w: room is unavailable", ErrConflict)
	}
	var overlapping int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM foster_orders WHERE room_id=? AND status IN ('PENDING','CONFIRMED','IN_PROGRESS') AND start_time < ? AND end_time > ?`, input.RoomID, formatStoredTime(input.EndTime), formatStoredTime(input.StartTime)).Scan(&overlapping)
	if err != nil {
		return FosterOrder{}, err
	}
	if overlapping >= capacity {
		return FosterOrder{}, fmt.Errorf("%w: room capacity has been reached", ErrConflict)
	}
	days := math.Ceil(input.EndTime.Sub(input.StartTime).Hours() / 24)
	if days < 1 {
		days = 1
	}
	total := price * days
	serviceRows := make([]OrderService, 0, len(input.Services))
	for _, selected := range input.Services {
		if selected.Quantity < 1 {
			selected.Quantity = 1
		}
		var servicePrice float64
		var active int
		if err := tx.QueryRowContext(ctx, `SELECT price,status FROM service_items WHERE service_id=?`, selected.ServiceID).Scan(&servicePrice, &active); errors.Is(err, sql.ErrNoRows) {
			return FosterOrder{}, ErrNotFound
		} else if err != nil {
			return FosterOrder{}, err
		}
		if active != 1 {
			return FosterOrder{}, fmt.Errorf("%w: selected service is unavailable", ErrConflict)
		}
		selected.Subtotal = servicePrice * float64(selected.Quantity)
		total += selected.Subtotal
		serviceRows = append(serviceRows, selected)
	}
	now := s.now().UTC()
	orderNo := fmt.Sprintf("FO%s%03d", now.Format("20060102150405"), now.Nanosecond()/1_000_000)
	result, err := tx.ExecContext(ctx, `INSERT INTO foster_orders(order_no,pet_id,user_id,room_id,start_time,end_time,room_type,total_amount,status,remarks,create_time,update_time) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, orderNo, input.PetID, ownerID, input.RoomID, formatStoredTime(input.StartTime), formatStoredTime(input.EndTime), roomType, total, "PENDING", input.Remarks, formatStoredTime(now), formatStoredTime(now))
	if err != nil {
		return FosterOrder{}, translateServiceError(err)
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		return FosterOrder{}, err
	}
	for _, selected := range serviceRows {
		if _, err := tx.ExecContext(ctx, `INSERT INTO order_services(order_id,service_id,quantity,subtotal,create_time) VALUES(?,?,?,?,?)`, orderID, selected.ServiceID, selected.Quantity, selected.Subtotal, formatStoredTime(now)); err != nil {
			return FosterOrder{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE rooms SET status='RESERVED',update_time=? WHERE room_id=? AND status='AVAILABLE'`, formatStoredTime(now), input.RoomID); err != nil {
		return FosterOrder{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO operation_logs(user_id,username,operation,method,result,create_time) VALUES(?,?,?,?,1,?)`, principal.UserID, principal.Username, "CREATE_ORDER", "POST /api/order/create", formatStoredTime(now)); err != nil {
		return FosterOrder{}, err
	}
	if err := tx.Commit(); err != nil {
		return FosterOrder{}, err
	}
	return s.GetOrder(ctx, principal, orderID)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, principal Principal, id int64, next string) error {
	if principal.Role != RoleAdmin {
		return ErrForbidden
	}
	tx, err := s.store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	var roomID int64
	if err := tx.QueryRowContext(ctx, `SELECT status,COALESCE(room_id,0) FROM foster_orders WHERE order_id=?`, id).Scan(&current, &roomID); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if !validOrderTransition(current, next) {
		return fmt.Errorf("%w: %s cannot transition to %s", ErrInvalidState, current, next)
	}
	now := formatStoredTime(s.now())
	if _, err := tx.ExecContext(ctx, `UPDATE foster_orders SET status=?,update_time=? WHERE order_id=? AND status=?`, next, now, id, current); err != nil {
		return err
	}
	roomStatus := "RESERVED"
	switch next {
	case "IN_PROGRESS":
		roomStatus = "OCCUPIED"
	case "COMPLETED", "CANCELLED":
		roomStatus = "CLEANING"
	}
	if roomID != 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE rooms SET status=?,update_time=? WHERE room_id=?`, roomStatus, now, roomID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO operation_logs(user_id,username,operation,method,result,create_time) VALUES(?,?,?,?,1,?)`, principal.UserID, principal.Username, "ORDER_STATUS_"+next, "PUT /api/order/status", now); err != nil {
		return err
	}
	return tx.Commit()
}
func validOrderTransition(current, next string) bool {
	switch current {
	case "PENDING":
		return next == "CONFIRMED" || next == "CANCELLED"
	case "CONFIRMED":
		return next == "IN_PROGRESS" || next == "CANCELLED"
	case "IN_PROGRESS":
		return next == "COMPLETED"
	default:
		return false
	}
}
func (s *Service) CancelOrder(ctx context.Context, p Principal, id int64) error {
	order, err := s.GetOrder(ctx, p, id)
	if err != nil {
		return err
	}
	if p.Role != RoleAdmin && order.UserID != p.UserID {
		return ErrForbidden
	}
	if order.Status != "PENDING" && order.Status != "CONFIRMED" {
		return fmt.Errorf("%w: only pending or confirmed orders can be cancelled", ErrInvalidState)
	}
	admin := p
	admin.Role = RoleAdmin
	return s.UpdateOrderStatus(ctx, admin, id, "CANCELLED")
}
func (s *Service) GetOrder(ctx context.Context, p Principal, id int64) (FosterOrder, error) {
	if p.UserID == 0 {
		return FosterOrder{}, ErrUnauthenticated
	}
	var item FosterOrder
	var start, end, created, updated string
	err := s.store.db.QueryRowContext(ctx, `SELECT o.order_id,o.order_no,COALESCE(o.pet_id,0),COALESCE(p.pet_name,''),o.user_id,COALESCE(u.username,''),COALESCE(o.room_id,0),COALESCE(r.room_number,''),o.start_time,o.end_time,COALESCE(o.room_type,''),COALESCE(o.service_package,''),o.total_amount,o.status,COALESCE(o.remarks,''),o.create_time,o.update_time FROM foster_orders o LEFT JOIN pets p ON p.pet_id=o.pet_id LEFT JOIN pet_users u ON u.user_id=o.user_id LEFT JOIN rooms r ON r.room_id=o.room_id WHERE o.order_id=?`, id).Scan(&item.ID, &item.OrderNo, &item.PetID, &item.PetName, &item.UserID, &item.Username, &item.RoomID, &item.RoomNumber, &start, &end, &item.RoomType, &item.ServicePackage, &item.TotalAmount, &item.Status, &item.Remarks, &created, &updated)
	if err == sql.ErrNoRows {
		return FosterOrder{}, ErrNotFound
	}
	if err != nil {
		return FosterOrder{}, err
	}
	if p.Role != RoleAdmin && item.UserID != p.UserID {
		return FosterOrder{}, ErrForbidden
	}
	item.StartTime, _ = parseStoredTime(start)
	item.EndTime, _ = parseStoredTime(end)
	item.CreatedAt, _ = parseStoredTime(created)
	item.UpdatedAt, _ = parseStoredTime(updated)
	rows, err := s.store.db.QueryContext(ctx, `SELECT service_id,quantity,subtotal FROM order_services WHERE order_id=? ORDER BY id`, id)
	if err != nil {
		return FosterOrder{}, err
	}
	defer rows.Close()
	item.Services = []OrderService{}
	for rows.Next() {
		var svc OrderService
		if err := rows.Scan(&svc.ServiceID, &svc.Quantity, &svc.Subtotal); err != nil {
			return FosterOrder{}, err
		}
		item.Services = append(item.Services, svc)
	}
	return item, rows.Err()
}
func (s *Service) ListOrders(ctx context.Context, p Principal, pageNum, pageSize int, orderNo, petName, username, status string, userID int64) (Page[FosterOrder], error) {
	if p.UserID == 0 {
		return Page[FosterOrder]{}, ErrUnauthenticated
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
	if orderNo != "" {
		where = append(where, "o.order_no LIKE ?")
		args = append(args, "%"+orderNo+"%")
	}
	if petName != "" {
		where = append(where, "p.pet_name LIKE ?")
		args = append(args, "%"+petName+"%")
	}
	if username != "" {
		where = append(where, "u.username LIKE ?")
		args = append(args, "%"+username+"%")
	}
	if status != "" {
		where = append(where, "o.status=?")
		args = append(args, status)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := s.store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM foster_orders o LEFT JOIN pets p ON p.pet_id=o.pet_id LEFT JOIN pet_users u ON u.user_id=o.user_id WHERE "+clause, args...).Scan(&total); err != nil {
		return Page[FosterOrder]{}, err
	}
	query := "SELECT o.order_id,o.order_no,COALESCE(o.pet_id,0),COALESCE(p.pet_name,''),o.user_id,COALESCE(u.username,''),COALESCE(o.room_id,0),COALESCE(r.room_number,''),o.start_time,o.end_time,COALESCE(o.room_type,''),COALESCE(o.service_package,''),o.total_amount,o.status,COALESCE(o.remarks,''),o.create_time,o.update_time FROM foster_orders o LEFT JOIN pets p ON p.pet_id=o.pet_id LEFT JOIN pet_users u ON u.user_id=o.user_id LEFT JOIN rooms r ON r.room_id=o.room_id WHERE " + clause + " ORDER BY o.order_id DESC LIMIT ? OFFSET ?"
	rows, err := s.store.db.QueryContext(ctx, query, append(args, pageSize, (pageNum-1)*pageSize)...)
	if err != nil {
		return Page[FosterOrder]{}, err
	}
	defer rows.Close()
	items := make([]FosterOrder, 0)
	for rows.Next() {
		var item FosterOrder
		var start, end, created, updated string
		if err := rows.Scan(&item.ID, &item.OrderNo, &item.PetID, &item.PetName, &item.UserID, &item.Username, &item.RoomID, &item.RoomNumber, &start, &end, &item.RoomType, &item.ServicePackage, &item.TotalAmount, &item.Status, &item.Remarks, &created, &updated); err != nil {
			return Page[FosterOrder]{}, err
		}
		item.StartTime, _ = parseStoredTime(start)
		item.EndTime, _ = parseStoredTime(end)
		item.CreatedAt, _ = parseStoredTime(created)
		item.UpdatedAt, _ = parseStoredTime(updated)
		items = append(items, item)
	}
	return Page[FosterOrder]{List: items, Total: total, PageNum: pageNum, PageSize: pageSize}, rows.Err()
}
func (s *Service) OrderStatistics(ctx context.Context, p Principal) (map[string]any, error) {
	if p.UserID == 0 {
		return nil, ErrUnauthenticated
	}
	where := ""
	args := []any{}
	if p.Role != RoleAdmin {
		where = " WHERE user_id=?"
		args = append(args, p.UserID)
	}
	rows, err := s.store.db.QueryContext(ctx, "SELECT status,COUNT(*) FROM foster_orders"+where+" GROUP BY status", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := map[string]any{"total": 0, "pending": 0, "confirmed": 0, "inProgress": 0, "completed": 0, "cancelled": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats["total"] = stats["total"].(int) + count
		switch status {
		case "PENDING":
			stats["pending"] = count
		case "CONFIRMED":
			stats["confirmed"] = count
		case "IN_PROGRESS":
			stats["inProgress"] = count
		case "COMPLETED":
			stats["completed"] = count
		case "CANCELLED":
			stats["cancelled"] = count
		}
	}
	return stats, rows.Err()
}

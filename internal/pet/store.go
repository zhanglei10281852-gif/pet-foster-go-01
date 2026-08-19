package pet

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(ctx context.Context, path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("pet database path is required")
	}
	if path != ":memory:" && !strings.HasPrefix(path, "file:") {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create pet database directory: %w", err)
		}
	}
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	if strings.Contains(path, "?") {
		dsn = path + "&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open pet database: %w", err)
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	if path == ":memory:" || strings.Contains(path, "mode=memory") {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping pet database: %w", err)
	}
	store := &Store{db: db}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.seed(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error                   { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *Store) migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS pet_schema_versions (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS pet_users (user_id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, phone TEXT, email TEXT, address TEXT, role TEXT NOT NULL, status INTEGER NOT NULL DEFAULT 1, create_time TEXT NOT NULL, update_time TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS pet_sessions (session_id TEXT PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES pet_users(user_id) ON DELETE CASCADE, token_hash TEXT NOT NULL UNIQUE, expires_at TEXT NOT NULL, revoked_at TEXT, create_time TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS pets (pet_id INTEGER PRIMARY KEY AUTOINCREMENT, pet_name TEXT NOT NULL, pet_type TEXT NOT NULL, breed TEXT, age INTEGER, weight REAL, health_status TEXT, special_requirements TEXT, avatar TEXT, owner_id INTEGER NOT NULL REFERENCES pet_users(user_id) ON DELETE CASCADE, create_time TEXT NOT NULL, update_time TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS rooms (room_id INTEGER PRIMARY KEY AUTOINCREMENT, room_number TEXT NOT NULL UNIQUE, room_type TEXT NOT NULL, status TEXT NOT NULL, price_per_day REAL NOT NULL, description TEXT, capacity INTEGER NOT NULL DEFAULT 1, create_time TEXT NOT NULL, update_time TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS service_items (service_id INTEGER PRIMARY KEY AUTOINCREMENT, service_name TEXT NOT NULL, description TEXT, price REAL NOT NULL, status INTEGER NOT NULL DEFAULT 1, create_time TEXT NOT NULL, update_time TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS foster_orders (order_id INTEGER PRIMARY KEY AUTOINCREMENT, order_no TEXT NOT NULL UNIQUE, pet_id INTEGER REFERENCES pets(pet_id) ON DELETE SET NULL, user_id INTEGER NOT NULL REFERENCES pet_users(user_id), room_id INTEGER REFERENCES rooms(room_id) ON DELETE SET NULL, start_time TEXT NOT NULL, end_time TEXT NOT NULL, room_type TEXT, service_package TEXT, total_amount REAL NOT NULL DEFAULT 0, status TEXT NOT NULL, remarks TEXT, create_time TEXT NOT NULL, update_time TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS order_services (id INTEGER PRIMARY KEY AUTOINCREMENT, order_id INTEGER NOT NULL REFERENCES foster_orders(order_id) ON DELETE CASCADE, service_id INTEGER NOT NULL REFERENCES service_items(service_id), quantity INTEGER NOT NULL, subtotal REAL NOT NULL, create_time TEXT NOT NULL, UNIQUE(order_id, service_id))`,
		`CREATE TABLE IF NOT EXISTS daily_records (record_id INTEGER PRIMARY KEY AUTOINCREMENT, order_id INTEGER NOT NULL REFERENCES foster_orders(order_id) ON DELETE CASCADE, record_date TEXT NOT NULL, diet TEXT, defecation TEXT, activity TEXT, spirit TEXT, remarks TEXT, media_urls TEXT, create_time TEXT NOT NULL, update_time TEXT NOT NULL, UNIQUE(order_id, record_date))`,
		`CREATE TABLE IF NOT EXISTS operation_logs (log_id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, username TEXT, operation TEXT NOT NULL, method TEXT, params TEXT, ip TEXT, result INTEGER, error_msg TEXT, cost_time INTEGER, create_time TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_pet_owner ON pets(owner_id)`,
		`CREATE INDEX IF NOT EXISTS idx_order_user_status ON foster_orders(user_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_order_time ON foster_orders(start_time, end_time)`,
		`CREATE INDEX IF NOT EXISTS idx_record_order_date ON daily_records(order_id, record_date)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("pet migration: %w", err)
		}
	}
	return nil
}

func (s *Store) seed(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pet_users`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		adminHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		userHash, err := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		for _, row := range []struct{ username, hash, role string }{{"admin", string(adminHash), "ADMIN"}, {"testuser", string(userHash), "USER"}} {
			if _, err := s.db.ExecContext(ctx, `INSERT INTO pet_users(username,password_hash,phone,email,address,role,status,create_time,update_time) VALUES(?,?,?,?,?,?,1,?,?)`, row.username, row.hash, "", "", "", row.role, now, now); err != nil {
				return err
			}
		}
	}
	if err := s.seedRooms(ctx); err != nil {
		return err
	}
	return s.seedServices(ctx)
}

func (s *Store) seedRooms(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM rooms`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	rooms := []struct {
		number, typ, status, description string
		price                            float64
		capacity                         int
	}{
		{"A101", "STANDARD", "AVAILABLE", "标准单间", 80, 1}, {"A102", "STANDARD", "AVAILABLE", "标准单间", 80, 1}, {"B201", "DELUXE", "AVAILABLE", "宽敞寄养间", 150, 2}, {"C301", "VIP", "AVAILABLE", "独立活动区", 280, 3},
	}
	for _, room := range rooms {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO rooms(room_number,room_type,status,price_per_day,description,capacity,create_time,update_time) VALUES(?,?,?,?,?,?,?,?)`, room.number, room.typ, room.status, room.price, room.description, room.capacity, now, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) seedServices(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM service_items`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	services := []struct {
		name, description string
		price             float64
	}{
		{"基础护理", "每日梳毛与清洁耳朵", 30}, {"洗浴美容", "专业洗浴与吹干", 80}, {"遛狗服务", "每日两次户外活动", 50}, {"健康检查", "体温与体重监测", 20}, {"视频直播", "全天查看寄养状态", 15}, {"特殊饮食", "按宠物需求定制喂食", 40},
	}
	for _, item := range services {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO service_items(service_name,description,price,status,create_time,update_time) VALUES(?,?,?,1,?,?)`, item.name, item.description, item.price, now, now); err != nil {
			return err
		}
	}
	return nil
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func parseStoredTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stored time: %w", err)
	}
	return parsed.UTC(), nil
}

func formatStoredTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

package pet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUnauthenticated = errors.New("pet authentication required")
	ErrForbidden       = errors.New("pet permission denied")
	ErrNotFound        = errors.New("pet resource not found")
	ErrConflict        = errors.New("pet resource conflict")
	ErrValidation      = errors.New("pet request validation failed")
	ErrInvalidState    = errors.New("pet invalid state transition")
)

type Service struct {
	store *Store
	now   func() time.Time
}

func NewService(store *Store) *Service { return &Service{store: store, now: time.Now} }
func (s *Service) withNow(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *Service) Login(ctx context.Context, username, password string) (string, User, time.Time, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", User{}, time.Time{}, fmt.Errorf("%w: username and password are required", ErrValidation)
	}
	var id int64
	var hash, role, phone, email, address string
	var status int
	var created, updated string
	err := s.store.db.QueryRowContext(ctx, `SELECT user_id,password_hash,phone,email,address,role,status,create_time,update_time FROM pet_users WHERE username=?`, username).Scan(&id, &hash, &phone, &email, &address, &role, &status, &created, &updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", User{}, time.Time{}, fmt.Errorf("%w: invalid credentials", ErrUnauthenticated)
		}
		return "", User{}, time.Time{}, fmt.Errorf("load credentials: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil || status != 1 {
		return "", User{}, time.Time{}, fmt.Errorf("%w: invalid credentials", ErrUnauthenticated)
	}
	createdAt, err := parseStoredTime(created)
	if err != nil {
		return "", User{}, time.Time{}, err
	}
	updatedAt, err := parseStoredTime(updated)
	if err != nil {
		return "", User{}, time.Time{}, err
	}
	user := User{ID: id, Username: username, Phone: phone, Email: email, Address: address, Role: Role(role), Status: status, CreatedAt: createdAt, UpdatedAt: updatedAt}
	token, err := newToken()
	if err != nil {
		return "", User{}, time.Time{}, fmt.Errorf("create session token: %w", err)
	}
	expires := s.now().UTC().Add(24 * time.Hour)
	if _, err := s.store.db.ExecContext(ctx, `INSERT INTO pet_sessions(session_id,user_id,token_hash,expires_at,create_time) VALUES(?,?,?,?,?)`, token, id, tokenHash(token), formatStoredTime(expires), formatStoredTime(s.now())); err != nil {
		return "", User{}, time.Time{}, fmt.Errorf("create session: %w", err)
	}
	return token, user, expires, nil
}

func (s *Service) Logout(ctx context.Context, principal Principal) error {
	if principal.UserID == 0 {
		return ErrUnauthenticated
	}
	if _, err := s.store.db.ExecContext(ctx, `UPDATE pet_sessions SET revoked_at=? WHERE session_id=? AND revoked_at IS NULL`, formatStoredTime(s.now()), principal.SessionID); err != nil {
		return err
	}
	return nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (Principal, error) {
	if strings.TrimSpace(token) == "" {
		return Principal{}, ErrUnauthenticated
	}
	var sessionID, role, username, expires, revoked string
	var userID int64
	err := s.store.db.QueryRowContext(ctx, `SELECT s.session_id,s.user_id,u.username,u.role,s.expires_at,COALESCE(s.revoked_at,'') FROM pet_sessions s JOIN pet_users u ON u.user_id=s.user_id WHERE s.token_hash=?`, tokenHash(token)).Scan(&sessionID, &userID, &username, &role, &expires, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return Principal{}, ErrUnauthenticated
	}
	if err != nil {
		return Principal{}, err
	}
	expiresAt, err := parseStoredTime(expires)
	if err != nil {
		return Principal{}, err
	}
	if revoked != "" || !expiresAt.After(s.now()) {
		return Principal{}, ErrUnauthenticated
	}
	return Principal{UserID: userID, Username: username, Role: Role(role), SessionID: sessionID}, nil
}

func (s *Service) CurrentUser(ctx context.Context, principal Principal) (User, error) {
	if principal.UserID == 0 {
		return User{}, ErrUnauthenticated
	}
	return s.userByID(ctx, principal.UserID)
}

func (s *Service) userByID(ctx context.Context, id int64) (User, error) {
	var user User
	var role string
	var created, updated string
	var passwordHash string
	err := s.store.db.QueryRowContext(ctx, `SELECT user_id,username,password_hash,phone,email,address,role,status,create_time,update_time FROM pet_users WHERE user_id=?`, id).Scan(&user.ID, &user.Username, &passwordHash, &user.Phone, &user.Email, &user.Address, &role, &user.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	user.Role = Role(role)
	user.CreatedAt, err = parseStoredTime(created)
	if err != nil {
		return User{}, err
	}
	user.UpdatedAt, err = parseStoredTime(updated)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, principal Principal, input User) (User, error) {
	if principal.UserID == 0 {
		return User{}, ErrUnauthenticated
	}
	if input.ID == 0 {
		input.ID = principal.UserID
	}
	if input.ID != principal.UserID && principal.Role != RoleAdmin {
		return User{}, ErrForbidden
	}
	if strings.TrimSpace(input.Username) == "" {
		return User{}, fmt.Errorf("%w: username is required", ErrValidation)
	}
	role := input.Role
	if role == "" {
		role = RoleUser
	}
	if principal.Role != RoleAdmin {
		role = RoleUser
	}
	now := formatStoredTime(s.now())
	_, err := s.store.db.ExecContext(ctx, `UPDATE pet_users SET phone=?,email=?,address=?,role=?,status=?,update_time=? WHERE user_id=?`, input.Phone, input.Email, input.Address, role, input.Status, now, input.ID)
	if err != nil {
		return User{}, translateServiceError(err)
	}
	return s.userByID(ctx, input.ID)
}

func (s *Service) ListUsers(ctx context.Context, pageNum, pageSize int, username, role string) (Page[User], error) {
	pageNum, pageSize = normalizePage(pageNum, pageSize)
	where := []string{"1=1"}
	args := []any{}
	if strings.TrimSpace(username) != "" {
		where = append(where, "username LIKE ?")
		args = append(args, "%"+strings.TrimSpace(username)+"%")
	}
	if strings.TrimSpace(role) != "" {
		where = append(where, "role=?")
		args = append(args, role)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := s.store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pet_users WHERE "+clause, args...).Scan(&total); err != nil {
		return Page[User]{}, err
	}
	rows, err := s.store.db.QueryContext(ctx, "SELECT user_id,username,phone,email,address,role,status,create_time,update_time FROM pet_users WHERE "+clause+" ORDER BY user_id DESC LIMIT ? OFFSET ?", append(args, pageSize, (pageNum-1)*pageSize)...)
	if err != nil {
		return Page[User]{}, err
	}
	defer rows.Close()
	items := make([]User, 0)
	for rows.Next() {
		var item User
		var roleValue, created, updated string
		if err := rows.Scan(&item.ID, &item.Username, &item.Phone, &item.Email, &item.Address, &roleValue, &item.Status, &created, &updated); err != nil {
			return Page[User]{}, err
		}
		item.Role = Role(roleValue)
		item.CreatedAt, _ = parseStoredTime(created)
		item.UpdatedAt, _ = parseStoredTime(updated)
		items = append(items, item)
	}
	return Page[User]{List: items, Total: total, PageNum: pageNum, PageSize: pageSize}, rows.Err()
}

func (s *Service) DeleteUser(ctx context.Context, principal Principal, id int64) error {
	if principal.Role != RoleAdmin {
		return ErrForbidden
	}
	if id == principal.UserID {
		return fmt.Errorf("%w: current admin cannot be deleted", ErrConflict)
	}
	result, err := s.store.db.ExecContext(ctx, `DELETE FROM pet_users WHERE user_id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result, ErrNotFound)
}

func (s *Service) ChangePassword(ctx context.Context, principal Principal, oldPassword, newPassword string) error {
	if principal.UserID == 0 {
		return ErrUnauthenticated
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("%w: password must contain at least six characters", ErrValidation)
	}
	var hash string
	if err := s.store.db.QueryRowContext(ctx, `SELECT password_hash FROM pet_users WHERE user_id=?`, principal.UserID).Scan(&hash); err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(oldPassword)) != nil {
		return fmt.Errorf("%w: old password is incorrect", ErrValidation)
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.store.db.ExecContext(ctx, `UPDATE pet_users SET password_hash=?,update_time=? WHERE user_id=?`, string(newHash), formatStoredTime(s.now()), principal.UserID)
	return err
}

func (s *Service) ResetPassword(ctx context.Context, principal Principal, userID int64, password string) error {
	if principal.Role != RoleAdmin {
		return ErrForbidden
	}
	if len(password) < 6 {
		return fmt.Errorf("%w: password must contain at least six characters", ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	result, err := s.store.db.ExecContext(ctx, `UPDATE pet_users SET password_hash=?,update_time=? WHERE user_id=?`, string(hash), formatStoredTime(s.now()), userID)
	if err != nil {
		return err
	}
	return requireAffected(result, ErrNotFound)
}

func normalizePage(pageNum, pageSize int) (int, int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageNum, pageSize
}
func requireAffected(result sql.Result, notFound error) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notFound
	}
	return nil
}
func translateServiceError(err error) error {
	if strings.Contains(strings.ToLower(err.Error()), "unique") {
		return fmt.Errorf("%w: duplicate value", ErrConflict)
	}
	return err
}

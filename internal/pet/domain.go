package pet

import (
	"encoding/json"
	"fmt"
	"time"
)

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

type User struct {
	ID        int64     `json:"userId"`
	Username  string    `json:"username"`
	Phone     string    `json:"phone,omitempty"`
	Email     string    `json:"email,omitempty"`
	Address   string    `json:"address,omitempty"`
	Role      Role      `json:"role"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updateTime"`
}

type Pet struct {
	ID                  int64     `json:"petId"`
	Name                string    `json:"petName"`
	Type                string    `json:"petType"`
	Breed               string    `json:"breed,omitempty"`
	Age                 int       `json:"age,omitempty"`
	Weight              float64   `json:"weight,omitempty"`
	HealthStatus        string    `json:"healthStatus,omitempty"`
	SpecialRequirements string    `json:"specialRequirements,omitempty"`
	Avatar              string    `json:"avatar,omitempty"`
	OwnerID             int64     `json:"ownerId"`
	OwnerName           string    `json:"ownerName,omitempty"`
	CreatedAt           time.Time `json:"createTime"`
	UpdatedAt           time.Time `json:"updateTime"`
}

type Room struct {
	ID          int64     `json:"roomId"`
	Number      string    `json:"roomNumber"`
	Type        string    `json:"roomType"`
	Status      string    `json:"status"`
	PricePerDay float64   `json:"pricePerDay"`
	Description string    `json:"description,omitempty"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"createTime"`
	UpdatedAt   time.Time `json:"updateTime"`
}

type ServiceItem struct {
	ID          int64     `json:"serviceId"`
	Name        string    `json:"serviceName"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"createTime"`
	UpdatedAt   time.Time `json:"updateTime"`
}

type OrderService struct {
	ServiceID int64   `json:"serviceId"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

type FosterOrder struct {
	ID             int64          `json:"orderId"`
	OrderNo        string         `json:"orderNo"`
	PetID          int64          `json:"petId"`
	PetName        string         `json:"petName,omitempty"`
	UserID         int64          `json:"userId"`
	Username       string         `json:"username,omitempty"`
	RoomID         int64          `json:"roomId"`
	RoomNumber     string         `json:"roomNumber,omitempty"`
	StartTime      time.Time      `json:"startTime"`
	EndTime        time.Time      `json:"endTime"`
	RoomType       string         `json:"roomType,omitempty"`
	ServicePackage string         `json:"servicePackage,omitempty"`
	TotalAmount    float64        `json:"totalAmount"`
	Status         string         `json:"status"`
	Remarks        string         `json:"remarks,omitempty"`
	Services       []OrderService `json:"services,omitempty"`
	CreatedAt      time.Time      `json:"createTime"`
	UpdatedAt      time.Time      `json:"updateTime"`
}

type DailyRecord struct {
	ID         int64     `json:"recordId"`
	OrderID    int64     `json:"orderId"`
	RecordDate time.Time `json:"recordDate"`
	Diet       string    `json:"diet,omitempty"`
	Defecation string    `json:"defecation,omitempty"`
	Activity   string    `json:"activity,omitempty"`
	Spirit     string    `json:"spirit,omitempty"`
	Remarks    string    `json:"remarks,omitempty"`
	MediaURLs  string    `json:"mediaUrls,omitempty"`
	CreatedAt  time.Time `json:"createTime"`
	UpdatedAt  time.Time `json:"updateTime"`
}

// UnmarshalJSON accepts both the date-only value used by the admin form and
// the RFC3339 timestamp used by API clients and persisted responses.
func (r *DailyRecord) UnmarshalJSON(data []byte) error {
	type dailyRecordAlias DailyRecord
	var payload struct {
		RecordDate json.RawMessage `json:"recordDate"`
		*dailyRecordAlias
	}
	payload.dailyRecordAlias = (*dailyRecordAlias)(r)
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	if len(payload.RecordDate) == 0 || string(payload.RecordDate) == "null" {
		r.RecordDate = time.Time{}
		return nil
	}
	var value string
	if err := json.Unmarshal(payload.RecordDate, &value); err != nil {
		return fmt.Errorf("recordDate must be a string: %w", err)
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		r.RecordDate = parsed
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return fmt.Errorf("recordDate must be YYYY-MM-DD or RFC3339: %w", err)
	}
	r.RecordDate = parsed
	return nil
}

type Page[T any] struct {
	List     []T `json:"list"`
	Total    int `json:"total"`
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
}

type Principal struct {
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	Role      Role   `json:"role"`
	SessionID string `json:"-"`
}

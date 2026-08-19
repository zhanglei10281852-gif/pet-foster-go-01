package pet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDailyRecordAcceptsDateOnlyAndRFC3339Dates(t *testing.T) {
	var dateOnly DailyRecord
	if err := json.Unmarshal([]byte(`{"orderId":7,"recordDate":"2026-08-20","diet":"正常"}`), &dateOnly); err != nil {
		t.Fatalf("date-only record JSON error = %v", err)
	}
	if got := dateOnly.RecordDate.Format("2006-01-02"); got != "2026-08-20" {
		t.Fatalf("date-only record date = %s", got)
	}

	stamp := time.Date(2026, 8, 20, 8, 30, 0, 0, time.UTC)
	input, err := json.Marshal(struct {
		RecordDate time.Time `json:"recordDate"`
	}{RecordDate: stamp})
	if err != nil {
		t.Fatal(err)
	}
	var timestamp DailyRecord
	if err := json.Unmarshal(input, &timestamp); err != nil {
		t.Fatalf("RFC3339 record JSON error = %v", err)
	}
	if !timestamp.RecordDate.Equal(stamp) {
		t.Fatalf("RFC3339 record date = %s", timestamp.RecordDate)
	}
}

func testService(t *testing.T) (*Service, func()) {
	t.Helper()
	store, err := Open(context.Background(), "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store)
	now := time.Date(2026, 8, 20, 8, 0, 0, 0, time.UTC)
	service.withNow(func() time.Time { return now })
	return service, func() { _ = store.Close() }
}

func loginAs(t *testing.T, service *Service, username, password string) Principal {
	t.Helper()
	token, _, _, err := service.Login(context.Background(), username, password)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := service.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	return principal
}

func TestSessionLifecycleAndRoleAuthorization(t *testing.T) {
	service, closeStore := testService(t)
	defer closeStore()
	admin := loginAs(t, service, "admin", "admin123")
	userToken, _, _, err := service.Login(context.Background(), "testuser", "user123")
	if err != nil {
		t.Fatal(err)
	}
	user, err := service.Authenticate(context.Background(), userToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddRoom(context.Background(), user, Room{Number: "D101", Type: "STANDARD", Status: "AVAILABLE", PricePerDay: 88, Capacity: 1}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("user add room error = %v", err)
	}
	if _, err := service.AddRoom(context.Background(), admin, Room{Number: "D101", Type: "STANDARD", Status: "AVAILABLE", PricePerDay: 88, Capacity: 1}); err != nil {
		t.Fatal(err)
	}
	if err := service.Logout(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), userToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("authenticate revoked token error = %v", err)
	}
}

func TestOrderLifecyclePersistsRelatedEntities(t *testing.T) {
	service, closeStore := testService(t)
	defer closeStore()
	ctx := context.Background()
	admin := loginAs(t, service, "admin", "admin123")
	user := loginAs(t, service, "testuser", "user123")
	petItem, err := service.AddPet(ctx, user, Pet{Name: "豆豆", Type: "DOG", Breed: "柯基", HealthStatus: "健康"})
	if err != nil {
		t.Fatal(err)
	}
	rooms, err := service.AvailableRooms(ctx, user, "STANDARD")
	if err != nil || len(rooms) == 0 {
		t.Fatalf("rooms=%v err=%v", rooms, err)
	}
	services, err := service.AvailableServices(ctx, user)
	if err != nil || len(services) == 0 {
		t.Fatalf("services=%v err=%v", services, err)
	}
	start := time.Date(2026, 8, 21, 2, 0, 0, 0, time.UTC)
	order, err := service.CreateOrder(ctx, user, CreateOrderInput{PetID: petItem.ID, RoomID: rooms[0].ID, StartTime: start, EndTime: start.Add(48 * time.Hour), Services: []OrderService{{ServiceID: services[0].ID, Quantity: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != "PENDING" || order.TotalAmount <= 0 {
		t.Fatalf("order=%+v", order)
	}
	for _, state := range []string{"CONFIRMED", "IN_PROGRESS"} {
		if err := service.UpdateOrderStatus(ctx, admin, order.ID, state); err != nil {
			t.Fatal(err)
		}
	}
	record, err := service.AddRecord(ctx, admin, DailyRecord{OrderID: order.ID, RecordDate: start, Diet: "正常", Activity: "活跃", Spirit: "良好"})
	if err != nil {
		t.Fatal(err)
	}
	if record.ID == 0 {
		t.Fatal("record id missing")
	}
	if err := service.UpdateOrderStatus(ctx, admin, order.ID, "COMPLETED"); err != nil {
		t.Fatal(err)
	}
	reloaded, err := service.GetOrder(ctx, user, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Status != "COMPLETED" {
		t.Fatalf("status=%s", reloaded.Status)
	}
}

func TestConcurrentRoomCapacityIsNotOversold(t *testing.T) {
	service, closeStore := testService(t)
	defer closeStore()
	ctx := context.Background()
	user := loginAs(t, service, "testuser", "user123")
	room, err := service.AddRoom(ctx, loginAs(t, service, "admin", "admin123"), Room{Number: "RACE-1", Type: "STANDARD", Status: "AVAILABLE", PricePerDay: 100, Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	pets := make([]Pet, 2)
	for i := range pets {
		pets[i], err = service.AddPet(ctx, user, Pet{Name: fmt.Sprintf("并发宠物%d", i), Type: "DOG"})
		if err != nil {
			t.Fatal(err)
		}
	}
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	gate := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	results := make(chan error, 2)
	for i := range pets {
		go func(index int) {
			defer wg.Done()
			<-gate
			_, err := service.CreateOrder(ctx, user, CreateOrderInput{PetID: pets[index].ID, RoomID: room.ID, StartTime: start, EndTime: start.Add(24 * time.Hour)})
			results <- err
		}(i)
	}
	close(gate)
	wg.Wait()
	close(results)
	success := 0
	conflict := 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrConflict) {
			conflict++
		} else {
			t.Fatalf("unexpected err=%v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}

func TestRestartRestoresPetData(t *testing.T) {
	path := t.TempDir() + "/pet.db"
	ctx := context.Background()
	store, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store)
	user := loginAs(t, service, "testuser", "user123")
	item, err := service.AddPet(ctx, user, Pet{Name: "重启恢复", Type: "CAT"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service = NewService(store)
	user = loginAs(t, service, "testuser", "user123")
	restored, err := service.GetPet(ctx, user, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Name != "重启恢复" {
		t.Fatalf("restored=%+v", restored)
	}
}

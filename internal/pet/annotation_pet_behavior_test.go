package pet

import (
	"context"
	"errors"
	"testing"
)

func TestLogoutRevokesOnlyCurrentSession(t *testing.T) {
	service, closeStore := testService(t)
	defer closeStore()
	ctx := context.Background()
	firstToken, _, _, err := service.Login(ctx, "testuser", "user123")
	if err != nil {
		t.Fatal(err)
	}
	secondToken, _, _, err := service.Login(ctx, "testuser", "user123")
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.Authenticate(ctx, firstToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(ctx, secondToken); err != nil {
		t.Fatal(err)
	}
	if err := service.Logout(ctx, first); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(ctx, firstToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("first session error = %v", err)
	}
	if _, err := service.Authenticate(ctx, secondToken); err != nil {
		t.Fatalf("second session error = %v", err)
	}
}

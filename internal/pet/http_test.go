package pet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPFrontendContract(t *testing.T) {
	service, closeStore := testService(t)
	defer closeStore()
	server := httptest.NewServer(NewHandler(service))
	defer server.Close()
	client := server.Client()
	body := bytes.NewBufferString(`{"username":"testuser","password":"user123"}`)
	response, err := client.Post(server.URL+"/api/user/login", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var login struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
			User  User   `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&login); err != nil {
		t.Fatal(err)
	}
	if login.Code != 200 || login.Data.Token == "" || login.Data.User.Role != RoleUser {
		t.Fatalf("login=%+v", login)
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/api/pet/list?pageNum=1&pageSize=10", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+login.Data.Token)
	response, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var page struct {
		Code int       `json:"code"`
		Data Page[Pet] `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&page); err != nil {
		t.Fatal(err)
	}
	if page.Code != 200 || page.Data.List == nil {
		t.Fatalf("page=%+v", page)
	}
}

func TestHTTPRejectsUnauthenticatedRequests(t *testing.T) {
	service, closeStore := testService(t)
	defer closeStore()
	request := httptest.NewRequest(http.MethodGet, "/api/order/list", nil)
	response := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

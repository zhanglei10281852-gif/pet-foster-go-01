package pet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type principalKey struct{}
type Handler struct{ service *Service }

func NewHandler(service *Service) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("POST /api/user/login", h.login)
	mux.Handle("/api/", h.auth(http.HandlerFunc(h.api)))
	return mux
}
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeEnvelope(w, http.StatusOK, "ok", map[string]string{"status": "ok"})
}
func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.service.store.Ping(r.Context()); err != nil {
		writeAPIError(w, err)
		return
	}
	writeEnvelope(w, http.StatusOK, "ready", map[string]string{"status": "ready"})
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &input); err != nil {
		writeAPIError(w, err)
		return
	}
	token, user, expires, err := h.service.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeEnvelope(w, http.StatusOK, "登录成功", map[string]any{"token": token, "expiresAt": expires, "user": user})
}
func (h *Handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			writeAPIError(w, ErrUnauthenticated)
			return
		}
		principal, err := h.service.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeAPIError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, principal)))
	})
}
func principalFrom(r *http.Request) Principal {
	value, _ := r.Context().Value(principalKey{}).(Principal)
	return value
}
func (h *Handler) api(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	p := principalFrom(r)
	switch {
	case r.Method == "POST" && path == "user/logout":
		err := h.service.Logout(r.Context(), p)
		h.finish(w, err, nil)
	case r.Method == "GET" && path == "user/info":
		data, err := h.service.CurrentUser(r.Context(), p)
		h.finish(w, err, data)
	case r.Method == "PUT" && path == "user/update":
		var input User
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.UpdateUser(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "GET" && path == "user/list":
		data, err := h.service.ListUsers(r.Context(), queryInt(r, "pageNum"), queryInt(r, "pageSize"), r.URL.Query().Get("username"), r.URL.Query().Get("role"))
		h.finish(w, err, data)
	case r.Method == "DELETE" && strings.HasPrefix(path, "user/"):
		h.finish(w, h.service.DeleteUser(r.Context(), p, pathID(path)), nil)
	case r.Method == "PUT" && path == "user/password":
		var input struct {
			OldPassword string `json:"oldPassword"`
			NewPassword string `json:"newPassword"`
		}
		err := decode(r, &input)
		if err == nil {
			err = h.service.ChangePassword(r.Context(), p, input.OldPassword, input.NewPassword)
		}
		h.finish(w, err, nil)
	case r.Method == "PUT" && path == "user/reset-password":
		h.finish(w, h.service.ResetPassword(r.Context(), p, queryInt64(r, "userId"), r.URL.Query().Get("newPassword")), nil)
	case r.Method == "POST" && path == "pet/add":
		var input Pet
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.AddPet(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "PUT" && path == "pet/update":
		var input Pet
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.UpdatePet(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "GET" && path == "pet/my":
		data, err := h.service.MyPets(r.Context(), p)
		h.finish(w, err, data)
	case r.Method == "GET" && path == "pet/list":
		data, err := h.service.ListPets(r.Context(), p, queryInt(r, "pageNum"), queryInt(r, "pageSize"), r.URL.Query().Get("petName"), r.URL.Query().Get("petType"), queryInt64(r, "ownerId"))
		h.finish(w, err, data)
	case strings.HasPrefix(path, "pet/"):
		id := pathID(path)
		if r.Method == "GET" {
			data, err := h.service.GetPet(r.Context(), p, id)
			h.finish(w, err, data)
		} else if r.Method == "DELETE" {
			h.finish(w, h.service.DeletePet(r.Context(), p, id), nil)
		} else {
			writeAPIError(w, ErrNotFound)
		}
	case r.Method == "POST" && path == "room/add":
		var input Room
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.AddRoom(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "PUT" && path == "room/update":
		var input Room
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.UpdateRoom(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "PUT" && path == "room/status":
		h.finish(w, h.service.UpdateRoomStatus(r.Context(), p, queryInt64(r, "roomId"), r.URL.Query().Get("status")), nil)
	case r.Method == "GET" && path == "room/available":
		data, err := h.service.AvailableRooms(r.Context(), p, r.URL.Query().Get("roomType"))
		h.finish(w, err, data)
	case r.Method == "GET" && path == "room/list":
		data, err := h.service.ListRooms(r.Context(), p, queryInt(r, "pageNum"), queryInt(r, "pageSize"), r.URL.Query().Get("roomNumber"), r.URL.Query().Get("roomType"), r.URL.Query().Get("status"))
		h.finish(w, err, data)
	case strings.HasPrefix(path, "room/"):
		id := pathID(path)
		if r.Method == "GET" {
			data, err := h.service.GetRoom(r.Context(), p, id)
			h.finish(w, err, data)
		} else if r.Method == "DELETE" {
			h.finish(w, h.service.DeleteRoom(r.Context(), p, id), nil)
		} else {
			writeAPIError(w, ErrNotFound)
		}
	case r.Method == "POST" && path == "service/add":
		var input ServiceItem
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.AddServiceItem(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "PUT" && path == "service/update":
		var input ServiceItem
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.UpdateServiceItem(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "GET" && path == "service/all":
		data, err := h.service.AllServices(r.Context(), p)
		h.finish(w, err, data)
	case r.Method == "GET" && path == "service/available":
		data, err := h.service.AvailableServices(r.Context(), p)
		h.finish(w, err, data)
	case r.Method == "GET" && path == "service/list":
		var status *int
		if raw := r.URL.Query().Get("status"); raw != "" {
			value, _ := strconv.Atoi(raw)
			status = &value
		}
		data, err := h.service.ListServiceItems(r.Context(), p, queryInt(r, "pageNum"), queryInt(r, "pageSize"), r.URL.Query().Get("serviceName"), status)
		h.finish(w, err, data)
	case strings.HasPrefix(path, "service/"):
		id := pathID(path)
		if r.Method == "GET" {
			data, err := h.service.GetServiceItem(r.Context(), p, id)
			h.finish(w, err, data)
		} else if r.Method == "DELETE" {
			h.finish(w, h.service.DeleteServiceItem(r.Context(), p, id), nil)
		} else {
			writeAPIError(w, ErrNotFound)
		}
	case r.Method == "POST" && path == "order/create":
		var input CreateOrderInput
		err := decode(r, &input)
		var data FosterOrder
		if err == nil {
			data, err = h.service.CreateOrder(r.Context(), p, input)
		}
		h.finish(w, err, data)
	case r.Method == "PUT" && path == "order/status":
		h.finish(w, h.service.UpdateOrderStatus(r.Context(), p, queryInt64(r, "orderId"), r.URL.Query().Get("status")), nil)
	case r.Method == "GET" && path == "order/statistics":
		data, err := h.service.OrderStatistics(r.Context(), p)
		h.finish(w, err, data)
	case r.Method == "GET" && (path == "order/list" || path == "order/my"):
		data, err := h.service.ListOrders(r.Context(), p, queryInt(r, "pageNum"), queryInt(r, "pageSize"), r.URL.Query().Get("orderNo"), r.URL.Query().Get("petName"), r.URL.Query().Get("username"), r.URL.Query().Get("status"), queryInt64(r, "userId"))
		h.finish(w, err, data)
	case strings.HasPrefix(path, "order/"):
		id := pathID(path)
		if r.Method == "GET" {
			data, err := h.service.GetOrder(r.Context(), p, id)
			h.finish(w, err, data)
		} else if r.Method == "DELETE" {
			h.finish(w, h.service.CancelOrder(r.Context(), p, id), nil)
		} else {
			writeAPIError(w, ErrNotFound)
		}
	case r.Method == "POST" && path == "record/add":
		var input DailyRecord
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.AddRecord(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "PUT" && path == "record/update":
		var input DailyRecord
		err := decode(r, &input)
		if err == nil {
			input, err = h.service.UpdateRecord(r.Context(), p, input)
		}
		h.finish(w, err, input)
	case r.Method == "GET" && strings.HasPrefix(path, "record/order/"):
		data, err := h.service.RecordsByOrder(r.Context(), p, pathID(path))
		h.finish(w, err, data)
	case r.Method == "GET" && path == "record/list":
		data, err := h.service.ListRecords(r.Context(), p, queryInt(r, "pageNum"), queryInt(r, "pageSize"), queryInt64(r, "orderId"), queryInt64(r, "userId"), r.URL.Query().Get("startDate"), r.URL.Query().Get("endDate"))
		h.finish(w, err, data)
	case strings.HasPrefix(path, "record/"):
		id := pathID(path)
		if r.Method == "GET" {
			data, err := h.service.GetRecord(r.Context(), p, id)
			h.finish(w, err, data)
		} else if r.Method == "DELETE" {
			h.finish(w, h.service.DeleteRecord(r.Context(), p, id), nil)
		} else {
			writeAPIError(w, ErrNotFound)
		}
	default:
		writeAPIError(w, ErrNotFound)
	}
}
func (h *Handler) finish(w http.ResponseWriter, err error, data any) {
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeEnvelope(w, http.StatusOK, "操作成功", data)
}
func decode(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: invalid JSON", ErrValidation)
	}
	return nil
}
func writeEnvelope(w http.ResponseWriter, status int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "message": message, "data": data})
}
func writeAPIError(w http.ResponseWriter, err error) {
	status, code, message := http.StatusInternalServerError, 500, "服务内部错误"
	switch {
	case errors.Is(err, ErrUnauthenticated):
		status, code, message = http.StatusUnauthorized, 401, "登录已过期，请重新登录"
	case errors.Is(err, ErrForbidden):
		status, code, message = http.StatusForbidden, 403, "没有权限执行此操作"
	case errors.Is(err, ErrNotFound):
		status, code, message = http.StatusNotFound, 404, "资源不存在"
	case errors.Is(err, ErrConflict), errors.Is(err, ErrInvalidState):
		status, code, message = http.StatusConflict, 409, err.Error()
	case errors.Is(err, ErrValidation):
		status, code, message = http.StatusBadRequest, 400, err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": message, "data": nil})
}
func queryInt(r *http.Request, key string) int {
	value, _ := strconv.Atoi(r.URL.Query().Get(key))
	return value
}
func queryInt64(r *http.Request, key string) int64 {
	value, _ := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	return value
}
func pathID(path string) int64 {
	value := path[strings.LastIndex(path, "/")+1:]
	id, _ := strconv.ParseInt(value, 10, 64)
	return id
}

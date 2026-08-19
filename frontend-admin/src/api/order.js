import request from "./request";

export function createOrder(data) {
  return request.post("/order/create", data);
}

export function updateOrder(data) {
  return request.put("/order/update", data);
}

export function updateOrderStatus(orderId, status) {
  return request.put("/order/status", null, { params: { orderId, status } });
}

export function cancelOrder(id) {
  return request.delete(`/order/${id}`);
}

export function getOrderById(id) {
  return request.get(`/order/${id}`);
}

export function getMyOrders(params) {
  return request.get("/order/my", { params });
}

export function getOrderList(params) {
  return request.get("/order/list", { params });
}

export function getStatistics() {
  return request.get("/order/statistics");
}

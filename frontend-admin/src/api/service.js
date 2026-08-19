import request from "./request";

export function addService(data) {
  return request.post("/service/add", data);
}

export function updateService(data) {
  return request.put("/service/update", data);
}

export function deleteService(id) {
  return request.delete(`/service/${id}`);
}

export function getServiceById(id) {
  return request.get(`/service/${id}`);
}

export function getAllServices() {
  return request.get("/service/all");
}

export function getAvailableServices() {
  return request.get("/service/available");
}

export function getServiceList(params) {
  return request.get("/service/list", { params });
}

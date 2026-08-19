import request from "./request";

export function addRecord(data) {
  return request.post("/record/add", data);
}

export function updateRecord(data) {
  return request.put("/record/update", data);
}

export function deleteRecord(id) {
  return request.delete(`/record/${id}`);
}

export function getRecordById(id) {
  return request.get(`/record/${id}`);
}

export function getRecordsByOrderId(orderId) {
  return request.get(`/record/order/${orderId}`);
}

export function getRecordList(params) {
  return request.get("/record/list", { params });
}

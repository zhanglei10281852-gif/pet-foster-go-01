import request from "./request";

export function addRoom(data) {
  return request.post("/room/add", data);
}

export function updateRoom(data) {
  return request.put("/room/update", data);
}

export function deleteRoom(id) {
  return request.delete(`/room/${id}`);
}

export function getRoomById(id) {
  return request.get(`/room/${id}`);
}

export function getAvailableRooms(roomType) {
  return request.get("/room/available", { params: { roomType } });
}

export function getRoomList(params) {
  return request.get("/room/list", { params });
}

export function updateRoomStatus(roomId, status) {
  return request.put("/room/status", null, { params: { roomId, status } });
}

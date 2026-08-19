import request from "./request";

export function login(data) {
  return request.post("/user/login", data);
}

export function register(data) {
  return request.post("/user/register", data);
}

export function getUserInfo() {
  return request.get("/user/info");
}

export function updateUser(data) {
  return request.put("/user/update", data);
}

export function getUserList(params) {
  return request.get("/user/list", { params });
}

export function deleteUser(id) {
  return request.delete(`/user/${id}`);
}

export function changePassword(data) {
  return request.put("/user/password", data);
}

export function resetPassword(userId, newPassword) {
  return request.put("/user/reset-password", null, {
    params: { userId, newPassword },
  });
}

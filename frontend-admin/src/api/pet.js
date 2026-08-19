import request from "./request";

export function addPet(data) {
  return request.post("/pet/add", data);
}

export function updatePet(data) {
  return request.put("/pet/update", data);
}

export function deletePet(id) {
  return request.delete(`/pet/${id}`);
}

export function getPetById(id) {
  return request.get(`/pet/${id}`);
}

export function getMyPets() {
  return request.get("/pet/my");
}

export function getPetList(params) {
  return request.get("/pet/list", { params });
}

import { defineStore } from "pinia";
import { login as loginApi, getUserInfo } from "@/api/user";

export const useUserStore = defineStore("user", {
  state: () => ({
    token: localStorage.getItem("token") || "",
    user: JSON.parse(localStorage.getItem("user") || "{}"),
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.user.role === "ADMIN",
    username: (state) => state.user.username || "",
  },

  actions: {
    async login(loginForm) {
      const res = await loginApi(loginForm);
      this.token = res.data.token;
      this.user = res.data.user;
      localStorage.setItem("token", res.data.token);
      localStorage.setItem("user", JSON.stringify(res.data.user));
      return res;
    },

    async fetchUserInfo() {
      const res = await getUserInfo();
      this.user = res.data;
      localStorage.setItem("user", JSON.stringify(res.data));
      return res;
    },

    logout() {
      this.token = "";
      this.user = {};
      localStorage.removeItem("token");
      localStorage.removeItem("user");
    },
  },
});

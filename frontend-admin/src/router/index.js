import { createRouter, createWebHistory } from "vue-router";

const routes = [
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/Login.vue"),
    meta: { title: "登录" },
  },
  {
    path: "/",
    component: () => import("@/views/Layout.vue"),
    redirect: "/dashboard",
    children: [
      {
        path: "dashboard",
        name: "Dashboard",
        component: () => import("@/views/Dashboard.vue"),
        meta: { title: "首页" },
      },
      {
        path: "user",
        name: "User",
        component: () => import("@/views/User.vue"),
        meta: { title: "用户管理", admin: true },
      },
      {
        path: "pet",
        name: "Pet",
        component: () => import("@/views/Pet.vue"),
        meta: { title: "宠物管理" },
      },
      {
        path: "room",
        name: "Room",
        component: () => import("@/views/Room.vue"),
        meta: { title: "房间管理", admin: true },
      },
      {
        path: "service",
        name: "Service",
        component: () => import("@/views/Service.vue"),
        meta: { title: "服务项目", admin: true },
      },
      {
        path: "order",
        name: "Order",
        component: () => import("@/views/Order.vue"),
        meta: { title: "寄养订单" },
      },
      {
        path: "record",
        name: "Record",
        component: () => import("@/views/Record.vue"),
        meta: { title: "日常记录" },
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem("token");
  const user = JSON.parse(localStorage.getItem("user") || "{}");

  if (to.path !== "/login" && !token) {
    next("/login");
  } else if (to.meta.admin && user.role !== "ADMIN") {
    next("/dashboard");
  } else {
    next();
  }
});

export default router;

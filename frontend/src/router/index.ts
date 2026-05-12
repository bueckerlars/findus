import { createRouter, createWebHistory } from "vue-router";
import { useSession } from "../session";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", component: () => import("../views/LoginView.vue"), meta: { guestOnly: true } },
    { path: "/register", component: () => import("../views/RegisterView.vue"), meta: { guestOnly: true } },
    { path: "/", component: () => import("../views/HomeView.vue") },
    { path: "/profile", component: () => import("../views/ProfileView.vue") },
    { path: "/locations", component: () => import("../views/LocationsListView.vue") },
    { path: "/locations/new", component: () => import("../views/LocationFormView.vue"), meta: { requiresAdmin: true } },
    { path: "/locations/:id", component: () => import("../views/LocationDetailView.vue") },
    { path: "/locations/:id/edit", component: () => import("../views/LocationFormView.vue"), meta: { requiresAdmin: true } },
    { path: "/items", component: () => import("../views/ItemsListView.vue") },
    { path: "/items/new", component: () => import("../views/ItemFormView.vue"), meta: { requiresAdmin: true } },
    { path: "/items/:id", component: () => import("../views/ItemDetailView.vue") },
    {
      path: "/items/:id/edit",
      meta: { requiresAdmin: true },
      redirect: (to) => ({
        path: `/items/${String(to.params.id)}`,
        query: { ...to.query, edit: "1" },
      }),
    },
    { path: "/search", component: () => import("../views/SearchView.vue") },
    { path: "/labels", component: () => import("../views/LabelsListView.vue") },
    { path: "/labels/new", component: () => import("../views/LabelFormView.vue"), meta: { requiresAdmin: true } },
    { path: "/labels/:id/edit", component: () => import("../views/LabelFormView.vue"), meta: { requiresAdmin: true } },
    { path: "/admin", redirect: "/admin/users" },
    { path: "/admin/users", component: () => import("../views/AdminUsersView.vue"), meta: { requiresAdmin: true } },
    { path: "/admin/templates", component: () => import("../views/AdminTemplatesView.vue"), meta: { requiresAdmin: true } },
    { path: "/admin/templates/new", component: () => import("../views/AdminTemplateFormView.vue"), meta: { requiresAdmin: true } },
    { path: "/admin/templates/:id/edit", component: () => import("../views/AdminTemplateFormView.vue"), meta: { requiresAdmin: true } },
  ],
});

router.beforeEach(async (to) => {
  const { user, refresh } = useSession();
  if (user.value === undefined) {
    await refresh();
  }
  const needAuth = !to.meta.guestOnly;
  if (needAuth && !user.value) {
    return { path: "/login", query: { next: to.fullPath } };
  }
  if (to.meta.requiresAdmin && user.value?.role !== "admin") {
    return { path: "/" };
  }
  if (to.meta.guestOnly && user.value) {
    return { path: "/" };
  }
  return true;
});

export { router };

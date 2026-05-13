import { createRouter, createWebHistory } from "vue-router";
import { useCreateModals } from "../composables/useCreateModals";
import { useSession } from "../session";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", component: () => import("../views/LoginView.vue"), meta: { guestOnly: true } },
    { path: "/register", component: () => import("../views/RegisterView.vue"), meta: { guestOnly: true } },
    {
      path: "/",
      component: () => import("../layouts/BaseLayout.vue"),
      children: [
        { path: "", component: () => import("../views/HomeView.vue") },
        { path: "profile", component: () => import("../views/ProfileView.vue") },
        { path: "locations", component: () => import("../views/LocationsListView.vue") },
        {
          path: "locations/new",
          meta: { requiresAdmin: true },
          redirect: (to) => {
            const { setPending } = useCreateModals();
            const pid = to.query.parent_id;
            setPending({
              kind: "location",
              parentId: typeof pid === "string" ? pid : undefined,
            });
            return { path: "/locations", replace: true };
          },
        },
        { path: "locations/:id", component: () => import("../views/LocationDetailView.vue") },
        { path: "locations/:id/edit", component: () => import("../views/LocationFormView.vue"), meta: { requiresAdmin: true } },
        { path: "items", component: () => import("../views/ItemsListView.vue") },
        {
          path: "items/new",
          meta: { requiresAdmin: true },
          redirect: (to) => {
            const { setPending } = useCreateModals();
            const loc = to.query.location_id;
            setPending({
              kind: "item",
              locationId: typeof loc === "string" ? loc : undefined,
            });
            return { path: "/items", replace: true };
          },
        },
        { path: "items/:id", component: () => import("../views/ItemDetailView.vue") },
        {
          path: "items/:id/edit",
          meta: { requiresAdmin: true },
          redirect: (to) => ({
            path: `/items/${String(to.params.id)}`,
            query: { ...to.query, edit: "1" },
          }),
        },
        { path: "search", component: () => import("../views/SearchView.vue") },
        { path: "labels", component: () => import("../views/LabelsListView.vue") },
        {
          path: "labels/new",
          meta: { requiresAdmin: true },
          redirect: () => {
            const { setPending } = useCreateModals();
            setPending({ kind: "label" });
            return { path: "/labels", replace: true };
          },
        },
        { path: "labels/:id/edit", component: () => import("../views/LabelFormView.vue"), meta: { requiresAdmin: true } },
        {
          path: "admin",
          component: () => import("../views/AdminLayout.vue"),
          meta: { requiresAdmin: true },
          redirect: "/admin/users",
          children: [
            { path: "users", component: () => import("../views/AdminUsersView.vue") },
            { path: "settings", component: () => import("../views/AdminSettingsView.vue") },
            { path: "label-generator", component: () => import("../views/AdminLabelGeneratorView.vue") },
            { path: "templates", component: () => import("../views/AdminTemplatesView.vue") },
            { path: "templates/new", component: () => import("../views/AdminTemplateFormView.vue") },
            { path: "templates/:id/edit", component: () => import("../views/AdminTemplateFormView.vue") },
          ],
        },
      ],
    },
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
  if (to.matched.some((r) => r.meta.requiresAdmin) && user.value?.role !== "admin") {
    return { path: "/" };
  }
  if (to.meta.guestOnly && user.value) {
    return { path: "/" };
  }
  return true;
});

export { router };

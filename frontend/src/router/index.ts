import { createRouter, createWebHistory } from "vue-router";
import { useCreateModals } from "../composables/useCreateModals";
import { canAccessAnyPermission, useSession } from "../session";
import { PERM_ITEMS_WRITE, PERM_LABELS_WRITE, PERM_LOCATIONS_WRITE } from "../permissions";

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
          meta: { requiresAnyPermission: [PERM_LOCATIONS_WRITE] },
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
        { path: "locations/:id/edit", component: () => import("../views/LocationFormView.vue"), meta: { requiresAnyPermission: [PERM_LOCATIONS_WRITE] } },
        { path: "items", component: () => import("../views/ItemsListView.vue") },
        {
          path: "items/new",
          meta: { requiresAnyPermission: [PERM_ITEMS_WRITE] },
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
          meta: { requiresAnyPermission: [PERM_ITEMS_WRITE] },
          redirect: (to) => ({
            path: `/items/${String(to.params.id)}`,
            query: { ...to.query, edit: "1" },
          }),
        },
        { path: "search", component: () => import("../views/SearchView.vue") },
        { path: "labels", component: () => import("../views/LabelsListView.vue") },
        {
          path: "labels/new",
          meta: { requiresAnyPermission: [PERM_LABELS_WRITE] },
          redirect: () => {
            const { setPending } = useCreateModals();
            setPending({ kind: "label" });
            return { path: "/labels", replace: true };
          },
        },
        { path: "labels/:id/edit", component: () => import("../views/LabelFormView.vue"), meta: { requiresAnyPermission: [PERM_LABELS_WRITE] } },
        {
          path: "admin",
          component: () => import("../views/AdminLayout.vue"),
          meta: { requiresAdmin: true },
          redirect: "/admin/users",
          children: [
            { path: "users", component: () => import("../views/AdminUsersView.vue") },
            { path: "groups", component: () => import("../views/AdminGroupsView.vue") },
            { path: "groups/new", component: () => import("../views/AdminGroupFormView.vue") },
            { path: "groups/:id/edit", component: () => import("../views/AdminGroupFormView.vue") },
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
  if (
    to.matched.some((r) => {
      const req = r.meta.requiresAnyPermission as string[] | undefined;
      return !!req?.length && !canAccessAnyPermission(req);
    })
  ) {
    return { path: "/" };
  }
  if (to.meta.guestOnly && user.value) {
    return { path: "/" };
  }
  return true;
});

export { router };

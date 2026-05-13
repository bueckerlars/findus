import { ref, computed } from "vue";
import type { User } from "./api";
import { api } from "./api";
import { applyFxTheme, persistLastUserFxTheme, resetFxTheme } from "./composables/useFxTheme";

const user = ref<User | null | undefined>(undefined);
const permissions = ref<string[]>([]);
const accessGroups = ref<{ id: string; name: string }[]>([]);

type MeResponse = {
  user: User;
  permissions?: string[];
  groups?: { id: string; name: string }[];
};

async function refresh() {
  try {
    const r = await api<MeResponse>("/api/me");
    user.value = r.user;
    permissions.value = r.permissions ?? [];
    accessGroups.value = r.groups ?? [];
  } catch {
    user.value = null;
    permissions.value = [];
    accessGroups.value = [];
  } finally {
    if (typeof document !== "undefined") {
      if (user.value) {
        applyFxTheme(user.value.theme);
        persistLastUserFxTheme(user.value.theme);
      } else {
        resetFxTheme();
      }
    }
  }
}

const isAdmin = computed(() => user.value?.role === "admin");

function can(perm: string): boolean {
  if (user.value?.role === "admin") return true;
  return permissions.value.includes(perm);
}

/** For router guards (non-Vue context). */
export function canAccessAnyPermission(required: string[]): boolean {
  if (!user.value) return false;
  if (user.value.role === "admin") return true;
  return required.some((p) => permissions.value.includes(p));
}

export function useSession() {
  return { user, isAdmin, permissions, accessGroups, can, refresh };
}

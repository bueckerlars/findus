import { ref, computed } from "vue";
import type { User } from "./api";
import { api } from "./api";
import { applyFxTheme, persistLastUserFxTheme, resetFxTheme } from "./composables/useFxTheme";

const user = ref<User | null | undefined>(undefined);

async function refresh() {
  try {
    const r = await api<{ user: User }>("/api/me");
    user.value = r.user;
  } catch {
    user.value = null;
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

export function useSession() {
  return { user, isAdmin, refresh };
}

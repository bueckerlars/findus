<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useSession } from "../session";
import { postJson } from "../api";
import { navActive } from "../utils/nav";
import { usernameInitial } from "../utils/initial";
import FxSvg from "./FxSvg.vue";
import { useCreateModals } from "../composables/useCreateModals";
import FxDropdownMenu from "./primitives/FxDropdownMenu.vue";
import FxDropdownItem from "./primitives/FxDropdownItem.vue";
import FxDropdownSeparator from "./primitives/FxDropdownSeparator.vue";

const props = withDefaults(
  defineProps<{
    mobileDrawerOpen?: boolean;
    mobileDrawerAsDialog?: boolean;
    mobileDrawerAriaHidden?: boolean;
  }>(),
  {
    mobileDrawerOpen: false,
    mobileDrawerAsDialog: false,
    mobileDrawerAriaHidden: false,
  },
);

const route = useRoute();
const router = useRouter();
const { user, isAdmin, refresh } = useSession();
const { openCreateItem, openCreateLocation, openCreateLabel } = useCreateModals();
void refresh();

const profilePhotoSrc = "/profile/photo";

const rootRef = ref<HTMLElement | null>(null);

const drawerMotionClass = computed(() =>
  props.mobileDrawerOpen
    ? "max-sm:translate-x-0 max-sm:pointer-events-auto"
    : "max-sm:-translate-x-full max-sm:pointer-events-none",
);

async function logout() {
  await postJson("/api/auth/logout", {});
  user.value = null;
  await router.push("/login");
}

function focusFirst() {
  const root = rootRef.value;
  if (!root) return;
  const el = root.querySelector<HTMLElement>(
    'button:not([disabled]), [href], input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
  );
  el?.focus();
}

defineExpose({ focusFirst });
</script>

<template>
  <aside
    id="fx-app-sidebar"
    ref="rootRef"
    :class="drawerMotionClass"
    class="flex h-[100dvh] flex-col border-r border-zinc-200/80 bg-white/95 backdrop-blur-md max-sm:fixed max-sm:inset-y-0 max-sm:left-0 max-sm:z-[40] max-sm:w-[min(18rem,85vw)] max-sm:max-w-[85vw] max-sm:shadow-xl max-sm:transition-transform max-sm:duration-200 max-sm:ease-out sm:sticky sm:top-0 sm:z-40 sm:w-52 sm:shrink-0 sm:translate-x-0 sm:pointer-events-auto sm:shadow-[4px_0_24px_-12px_rgba(15,23,42,0.12)]"
    :role="mobileDrawerAsDialog ? 'dialog' : undefined"
    :aria-modal="mobileDrawerAsDialog ? true : undefined"
    :aria-labelledby="mobileDrawerAsDialog ? 'fx-shell-drawer-title' : undefined"
    :aria-label="mobileDrawerAsDialog ? undefined : $t('common.mainNav')"
    :aria-hidden="mobileDrawerAriaHidden ? true : undefined"
  >
    <div class="flex shrink-0 items-center gap-2.5 border-b border-zinc-100/90 px-4 py-4">
      <RouterLink to="/" class="flex min-w-0 items-center gap-2.5 rounded-lg py-0.5 text-zinc-900 outline-offset-2 transition hover:text-sky-800">
        <span
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-bold text-white shadow-md shadow-sky-900/20 ring-1 ring-white/20"
          >F</span
        >
        <span id="fx-shell-drawer-title" class="truncate text-sm font-semibold tracking-tight">{{ $t("common.findus") }}</span>
      </RouterLink>
    </div>

    <div v-if="user && isAdmin" class="shrink-0 border-b border-zinc-100 px-3 py-3">
      <FxDropdownMenu align="center" :side-offset="4" content-class="min-w-[12rem]">
        <template #trigger>
          <button
            type="button"
            class="fx-btn-primary w-full"
            :aria-label="$t('nav.add')"
          >
            <FxSvg name="plus" class="h-4 w-4" />
            {{ $t("nav.add") }}
            <FxSvg name="chevronDown" class="h-4 w-4" />
          </button>
        </template>
        <FxDropdownItem icon="cube" @select="openCreateItem()">{{ $t("nav.newItem") }}</FxDropdownItem>
        <FxDropdownItem icon="mapPin" @select="openCreateLocation()">{{ $t("nav.newLocation") }}</FxDropdownItem>
        <FxDropdownItem icon="tag" @select="openCreateLabel()">{{ $t("nav.newLabel") }}</FxDropdownItem>
      </FxDropdownMenu>
    </div>

    <nav class="flex min-h-0 flex-1 flex-col gap-0.5 overflow-y-auto px-2 py-3" :aria-label="$t('common.sectionsNav')">
      <RouterLink to="/" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="home" /></span> {{ $t("common.home") }}
      </RouterLink>
      <RouterLink to="/locations" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/locations') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="mapPin" /></span> {{ $t("common.locations") }}
      </RouterLink>
      <RouterLink to="/items" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/items') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="cube" /></span> {{ $t("common.items") }}
      </RouterLink>
      <RouterLink to="/labels" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/labels') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="tag" /></span> {{ $t("common.labels") }}
      </RouterLink>
      <RouterLink to="/search" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/search') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="magnifyingGlass" /></span> {{ $t("common.search") }}
      </RouterLink>
      <RouterLink
        v-if="isAdmin"
        to="/tools"
        class="fx-sidebar-link mt-2"
        :class="{ 'fx-sidebar-link-active': navActive(route.path, '/tools') }"
      >
        <span class="fx-sidebar-link-icon"><FxSvg name="qr" /></span> {{ $t("common.tools") }}
      </RouterLink>
    </nav>

    <footer v-if="user" class="shrink-0 px-3 pb-3 pt-2">
      <FxDropdownMenu align="start" side="top" :side-offset="6" content-class="min-w-[12rem]">
        <template #trigger>
          <button
            type="button"
            class="flex w-full cursor-pointer items-center gap-2.5 rounded-lg border border-zinc-200/80 bg-white p-2 text-left shadow-card ring-1 ring-zinc-950/[0.03] transition-all duration-200 outline-offset-2 hover:border-zinc-300/90 hover:shadow-md focus-visible:outline focus-visible:ring-2 focus-visible:ring-sky-400/35"
            :aria-label="$t('common.accountMenu')"
            :title="user.username"
          >
            <img
              v-if="user.avatar_path"
              :src="profilePhotoSrc"
              alt=""
              class="h-8 w-8 shrink-0 rounded-full object-cover shadow-sm ring-2 ring-zinc-200/70"
            />
            <span
              v-else
              class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-sky-500 to-sky-700 text-xs font-semibold text-white shadow-sm ring-2 ring-white/90"
              >{{ usernameInitial(user.username) }}</span
            >
            <span class="min-w-0 flex-1 truncate text-sm font-semibold tracking-tight text-zinc-900">{{ user.username }}</span>
            <span class="shrink-0 text-zinc-400"><FxSvg name="chevronDown" class="h-4 w-4" /></span>
          </button>
        </template>
        <FxDropdownItem icon="users" to="/profile">{{ $t("common.profile") }}</FxDropdownItem>
        <FxDropdownItem v-if="isAdmin" icon="gear" to="/admin">{{ $t("common.administration") }}</FxDropdownItem>
        <FxDropdownSeparator />
        <FxDropdownItem tone="danger" @select="logout">{{ $t("common.logOut") }}</FxDropdownItem>
      </FxDropdownMenu>
    </footer>
  </aside>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useSession } from "../session";
import { postJson } from "../api";
import { navActive } from "../utils/nav";
import { usernameInitial } from "../utils/initial";
import FxSvg from "./FxSvg.vue";
import { useCreateModals } from "../composables/useCreateModals";

const route = useRoute();
const router = useRouter();
const { user, isAdmin, refresh } = useSession();
const { openCreateItem, openCreateLocation, openCreateLabel } = useCreateModals();
void refresh();

const profilePhotoSrc = "/profile/photo";
const addDetailsRef = ref<HTMLDetailsElement | null>(null);
const accountDetailsRef = ref<HTMLDetailsElement | null>(null);

function closeAddMenu() {
  const el = addDetailsRef.value;
  if (el) el.open = false;
}

function closeAccountMenu() {
  const el = accountDetailsRef.value;
  if (el) el.open = false;
}

async function logout() {
  await postJson("/api/auth/logout", {});
  user.value = null;
  await router.push("/login");
}
</script>

<template>
  <aside
    class="sticky top-0 z-40 flex h-[100dvh] w-56 shrink-0 flex-col border-r border-zinc-200/80 bg-white/95 shadow-[4px_0_24px_-12px_rgba(15,23,42,0.12)] backdrop-blur-md"
    aria-label="Main navigation"
  >
    <div class="flex shrink-0 items-center gap-3 border-b border-zinc-100/90 px-5 py-5">
      <RouterLink to="/" class="flex min-w-0 items-center gap-3 rounded-xl py-0.5 text-zinc-900 outline-offset-2 transition hover:text-sky-800">
        <span
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-bold text-white shadow-md shadow-sky-900/20 ring-1 ring-white/20"
          >F</span
        >
        <span class="truncate text-base font-semibold tracking-tight">Findus</span>
      </RouterLink>
    </div>
    <div v-if="user && isAdmin" class="shrink-0 border-b border-zinc-100 px-4 py-4">
      <details ref="addDetailsRef" class="group relative">
        <summary
          class="fx-btn-primary flex cursor-pointer list-none items-center justify-center gap-2 rounded-xl px-3 py-3 text-sm font-semibold shadow-md [&::-webkit-details-marker]:hidden"
        >
          <FxSvg name="plus" />
          Add
          <FxSvg name="chevronDown" />
        </summary>
        <div
          class="absolute left-4 right-4 z-30 mt-2 overflow-hidden rounded-xl border border-zinc-200/90 bg-white/95 py-1.5 shadow-xl shadow-zinc-900/10 ring-1 ring-zinc-950/5 backdrop-blur-md"
        >
          <button
            type="button"
            class="block w-full px-4 py-2.5 text-left text-sm font-medium text-zinc-800 transition hover:bg-sky-50/80 hover:text-sky-900"
            @click="
              closeAddMenu();
              openCreateItem();
            "
          >
            New item
          </button>
          <button
            type="button"
            class="block w-full px-4 py-2.5 text-left text-sm font-medium text-zinc-800 transition hover:bg-sky-50/80 hover:text-sky-900"
            @click="
              closeAddMenu();
              openCreateLocation();
            "
          >
            New location
          </button>
          <button
            type="button"
            class="block w-full px-4 py-2.5 text-left text-sm font-medium text-zinc-800 transition hover:bg-sky-50/80 hover:text-sky-900"
            @click="
              closeAddMenu();
              openCreateLabel();
            "
          >
            New label
          </button>
        </div>
      </details>
    </div>
    <nav class="flex min-h-0 flex-1 flex-col gap-0.5 overflow-y-auto px-3 py-5" aria-label="Sections">
      <RouterLink to="/" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="home" /></span> Home
      </RouterLink>
      <RouterLink to="/locations" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/locations') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="mapPin" /></span> Locations
      </RouterLink>
      <RouterLink to="/items" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/items') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="cube" /></span> Items
      </RouterLink>
      <RouterLink to="/labels" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/labels') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="tag" /></span> Labels
      </RouterLink>
      <RouterLink to="/search" class="fx-sidebar-link" :class="{ 'fx-sidebar-link-active': navActive(route.path, '/search') }">
        <span class="fx-sidebar-link-icon"><FxSvg name="magnifyingGlass" /></span> Search
      </RouterLink>
      <RouterLink
        v-if="isAdmin"
        to="/admin"
        class="fx-sidebar-link mt-3"
        :class="{ 'fx-sidebar-link-active': navActive(route.path, '/admin') }"
      >
        <span class="fx-sidebar-link-icon"><FxSvg name="gear" /></span> Admin
      </RouterLink>
    </nav>
    <footer v-if="user" class="shrink-0 px-3 pb-4 pt-2">
      <details ref="accountDetailsRef" class="group relative z-10">
        <summary
          class="flex w-full cursor-pointer list-none items-center gap-3 rounded-xl border border-zinc-200/80 bg-white p-3 text-left shadow-card ring-1 ring-zinc-950/[0.03] transition-all duration-200 outline-offset-2 hover:border-zinc-300/90 hover:shadow-md motion-reduce:transition-none [&::-webkit-details-marker]:hidden focus-visible:outline focus-visible:ring-2 focus-visible:ring-sky-400/35 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-50"
          :aria-label="'Account menu'"
          :title="user.username"
        >
          <img
            v-if="user.avatar_path"
            :src="profilePhotoSrc"
            alt=""
            class="h-10 w-10 shrink-0 rounded-full object-cover shadow-sm ring-2 ring-zinc-200/70"
          />
          <span
            v-else
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-semibold text-white shadow-md shadow-sky-900/15 ring-2 ring-white/90"
            >{{ usernameInitial(user.username) }}</span
          >
          <span class="min-w-0 flex-1 truncate text-sm font-semibold tracking-tight text-zinc-900">{{ user.username }}</span>
          <span class="shrink-0 text-zinc-400 transition-colors group-open:text-zinc-600"><FxSvg name="chevronDown" /></span>
        </summary>
        <div
          class="absolute bottom-full left-0 right-0 z-30 mb-2 overflow-hidden rounded-xl border border-zinc-200/90 bg-white/95 py-1.5 shadow-xl shadow-zinc-900/10 ring-1 ring-zinc-950/5 backdrop-blur-md"
        >
          <RouterLink
            to="/profile"
            class="block px-4 py-2.5 text-sm font-medium text-zinc-800 transition hover:bg-sky-50/80 hover:text-sky-900"
            @click="closeAccountMenu"
            >Profile</RouterLink
          >
          <div class="border-t border-zinc-100/90">
            <button
              type="button"
              class="w-full px-4 py-2.5 text-left text-sm font-medium text-red-700/90 transition hover:bg-red-50/90 hover:text-red-800"
              @click="closeAccountMenu(); logout()"
            >
              Log out
            </button>
          </div>
        </div>
      </details>
    </footer>
  </aside>
</template>

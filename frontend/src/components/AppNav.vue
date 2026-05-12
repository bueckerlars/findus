<script setup lang="ts">
import { RouterLink, useRouter } from "vue-router";
import { useSession } from "../session";
import { api } from "../api";

const { user, isAdmin, refresh } = useSession();
refresh();
const router = useRouter();

async function logout() {
  await api("/api/auth/logout", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: "{}",
  });
  user.value = null;
  router.push("/login");
}
</script>

<template>
  <aside
    class="sticky top-0 z-40 hidden h-screen w-64 shrink-0 flex-col border-r border-zinc-200/80 bg-white/90 shadow-nav backdrop-blur-md lg:flex"
    aria-label="Sidebar"
  >
    <div class="flex h-full flex-col px-3 py-6">
      <RouterLink to="/" class="mb-6 flex items-center gap-2 px-2">
        <span
          class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-bold text-white shadow-sm"
          >F</span
        >
        <span class="font-semibold tracking-tight text-zinc-900">Findus</span>
      </RouterLink>
      <nav class="flex flex-1 flex-col gap-1" aria-label="Main">
        <RouterLink to="/" class="fx-sidebar-link" active-class="fx-sidebar-link-active">Home</RouterLink>
        <RouterLink to="/locations" class="fx-sidebar-link" active-class="fx-sidebar-link-active">Locations</RouterLink>
        <RouterLink to="/items" class="fx-sidebar-link" active-class="fx-sidebar-link-active">Items</RouterLink>
        <RouterLink to="/search" class="fx-sidebar-link" active-class="fx-sidebar-link-active">Search</RouterLink>
        <RouterLink v-if="isAdmin" to="/labels" class="fx-sidebar-link" active-class="fx-sidebar-link-active"
          >Labels</RouterLink
        >
        <RouterLink v-if="isAdmin" to="/admin" class="fx-sidebar-link" active-class="fx-sidebar-link-active"
          >Admin</RouterLink
        >
      </nav>
      <div class="mt-auto border-t border-zinc-100 pt-4">
        <RouterLink to="/profile" class="fx-sidebar-link" active-class="fx-sidebar-link-active">Profile</RouterLink>
        <button type="button" class="fx-sidebar-link mt-1 w-full text-left text-red-600 hover:bg-red-50" @click="logout">
          Logout
        </button>
      </div>
    </div>
  </aside>
  <header
    class="sticky top-0 z-50 flex border-b border-zinc-200/80 bg-white/90 shadow-sm backdrop-blur-md lg:hidden"
  >
    <div class="mx-auto flex w-full max-w-5xl items-center justify-between gap-3 px-4 py-3">
      <RouterLink to="/" class="flex items-center gap-2">
        <span
          class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-bold text-white shadow-sm"
          >F</span
        >
        <span class="font-semibold">Findus</span>
      </RouterLink>
      <nav class="flex gap-2 text-sm">
        <RouterLink to="/items" class="text-sky-700">Items</RouterLink>
        <RouterLink to="/search" class="text-zinc-600">Search</RouterLink>
      </nav>
    </div>
  </header>
</template>

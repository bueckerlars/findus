<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import GuestNav from "../components/GuestNav.vue";
import { postJson } from "../api";
import { useSession } from "../session";

const route = useRoute();
const router = useRouter();
const { refresh } = useSession();

const username = ref("");
const password = ref("");
const err = ref("");

async function submit() {
  err.value = "";
  const next = typeof route.query.next === "string" ? route.query.next : "/";
  try {
    const r = await postJson<{ next: string }>("/api/auth/login", {
      username: username.value,
      password: password.value,
      next: next || "/",
    });
    await refresh();
    await router.push(r.next || "/");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Sign in failed";
  }
}
</script>

<template>
  <div class="min-h-full bg-zinc-100 font-sans text-zinc-900 antialiased fx-auth-shell">
    <GuestNav />
    <main class="mx-auto flex min-h-[calc(100vh-4rem)] max-w-md flex-col justify-center px-4 py-10 sm:px-6">
      <div class="fx-auth-card">
        <h1 class="text-center text-2xl font-semibold tracking-tight text-zinc-900">Welcome back</h1>
        <p class="mt-1 text-center text-sm text-zinc-500">Sign in to manage your inventory</p>
        <p v-if="err" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-center text-sm text-red-700">
          {{ err }}
        </p>
        <form class="mt-8 space-y-5" @submit.prevent="submit">
          <div>
            <label class="fx-label" for="username">Username</label>
            <input id="username" v-model="username" class="fx-input" required autocomplete="username" placeholder="Your username" />
          </div>
          <div>
            <label class="fx-label" for="password">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              class="fx-input"
              required
              autocomplete="current-password"
              placeholder="••••••••"
            />
          </div>
          <button type="submit" class="fx-btn-primary w-full">Sign in</button>
        </form>
      </div>
    </main>
  </div>
</template>

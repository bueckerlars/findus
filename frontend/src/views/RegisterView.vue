<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import GuestNav from "../components/GuestNav.vue";
import { api, postJson } from "../api";
import { useSession } from "../session";

type Bootstrap = {
  user_count: number;
  has_registration_mode: boolean;
  registration_mode: string;
};

const route = useRoute();
const router = useRouter();
const { refresh } = useSession();

const username = ref("");
const email = ref("");
const password = ref("");
const invite = ref(
  (typeof route.query.token === "string" && route.query.token) ||
    (typeof route.query.invite === "string" && route.query.invite) ||
    "",
);
const err = ref("");
const boot = ref<Bootstrap | null>(null);

const help = computed(() => {
  const b = boot.value;
  if (!b) return "";
  if (b.user_count === 0) return "You are creating the first admin account.";
  if (b.has_registration_mode && b.registration_mode === "invite") return "Register using your invite link.";
  if (b.has_registration_mode && b.registration_mode === "open") return "Open registration is enabled.";
  return "Create your Findus account.";
});

onMounted(async () => {
  try {
    boot.value = await api<Bootstrap>("/api/bootstrap");
  } catch {
    boot.value = { user_count: 0, has_registration_mode: false, registration_mode: "" };
  }
});

async function submit() {
  err.value = "";
  try {
    const r = await postJson<{ next: string }>("/api/auth/register", {
      username: username.value,
      email: email.value,
      password: password.value,
      invite: invite.value,
    });
    await refresh();
    await router.push(r.next || "/");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Registration failed";
  }
}
</script>

<template>
  <div class="min-h-full bg-zinc-100 font-sans text-zinc-900 antialiased fx-auth-shell">
    <GuestNav />
    <main class="mx-auto flex min-h-[calc(100vh-4rem)] max-w-md flex-col justify-center px-4 py-10 sm:px-6">
      <div class="fx-auth-card">
        <h1 class="text-center text-2xl font-semibold tracking-tight">Create account</h1>
        <p class="mt-2 text-center text-sm leading-relaxed text-zinc-500">{{ help }}</p>
        <p v-if="err" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-center text-sm text-red-700">
          {{ err }}
        </p>
        <form class="mt-8 space-y-5" @submit.prevent="submit">
          <div>
            <label class="fx-label" for="reg-user">Username</label>
            <input id="reg-user" v-model="username" class="fx-input" required placeholder="Choose a username" />
          </div>
          <div>
            <label class="fx-label" for="reg-email">Email</label>
            <input id="reg-email" v-model="email" type="email" class="fx-input" required placeholder="you@example.com" />
          </div>
          <div>
            <label class="fx-label" for="reg-pass">Password</label>
            <input id="reg-pass" v-model="password" type="password" class="fx-input" required placeholder="At least 10 characters" />
          </div>
          <button type="submit" class="fx-btn-primary w-full">Create account</button>
        </form>
      </div>
    </main>
  </div>
</template>

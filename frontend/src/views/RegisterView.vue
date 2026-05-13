<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import AuthBackground from "../components/AuthBackground.vue";
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
const { t } = useI18n();

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
  if (b.user_count === 0) return t("auth.register.helpFirstAdmin");
  if (b.has_registration_mode && b.registration_mode === "invite") return t("auth.register.helpInvite");
  if (b.has_registration_mode && b.registration_mode === "open") return t("auth.register.helpOpen");
  return t("auth.register.helpDefault");
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
    err.value = e instanceof Error ? e.message : t("common.registrationFailed");
  }
}
</script>

<template>
  <div class="fx-auth-shell min-h-full font-sans text-zinc-900 antialiased">
    <AuthBackground />
    <div class="relative z-[1] flex min-h-full flex-col">
      <GuestNav />
      <main class="mx-auto flex min-h-[calc(100vh-4rem)] max-w-md flex-1 flex-col justify-center px-4 py-10 sm:px-6">
        <div class="fx-auth-card">
          <h1 class="fx-auth-title">{{ $t("auth.register.title") }}</h1>
          <p class="mt-3 text-center text-base font-medium leading-relaxed text-zinc-600">{{ help }}</p>
          <p v-if="err" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-center text-sm text-red-700">
            {{ err }}
          </p>
          <form class="mt-8 space-y-5" @submit.prevent="submit">
            <div>
              <label class="fx-label" for="reg-user">{{ $t("auth.register.username") }}</label>
              <input id="reg-user" v-model="username" class="fx-input" required :placeholder="$t('auth.register.usernamePlaceholder')" />
            </div>
            <div>
              <label class="fx-label" for="reg-email">{{ $t("auth.register.email") }}</label>
              <input id="reg-email" v-model="email" type="email" class="fx-input" required :placeholder="$t('auth.register.emailPlaceholder')" />
            </div>
            <div>
              <label class="fx-label" for="reg-pass">{{ $t("auth.register.password") }}</label>
              <input id="reg-pass" v-model="password" type="password" class="fx-input" required :placeholder="$t('auth.register.passwordPlaceholder')" />
            </div>
            <button type="submit" class="fx-btn-primary w-full">{{ $t("auth.register.submit") }}</button>
          </form>
        </div>
      </main>
    </div>
  </div>
</template>

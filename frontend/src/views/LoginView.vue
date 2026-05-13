<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import AuthBackground from "../components/AuthBackground.vue";
import AuthFormInput from "../components/AuthFormInput.vue";
import GuestNav from "../components/GuestNav.vue";
import { postJson } from "../api";
import { useSession } from "../session";

const route = useRoute();
const router = useRouter();
const { refresh } = useSession();
const { t } = useI18n();

const username = ref("");
const password = ref("");
const err = ref("");
const submitting = ref(false);

async function submit() {
  if (submitting.value) return;
  err.value = "";
  const next = typeof route.query.next === "string" ? route.query.next : "/";
  submitting.value = true;
  try {
    const r = await postJson<{ next: string }>("/api/auth/login", {
      username: username.value,
      password: password.value,
      next: next || "/",
    });
    await refresh();
    await router.push(r.next || "/");
  } catch (e) {
    const raw = e instanceof Error ? e.message : "";
    const norm = raw.trim().toLowerCase();
    if (
      norm === "unauthorized" ||
      norm === "invalid credentials"
    ) {
      err.value = t("auth.login.invalidCredentials");
    } else {
      err.value = raw || t("common.signInFailed");
    }
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="fx-auth-shell min-h-full font-sans text-zinc-900 antialiased">
    <AuthBackground />
    <div class="relative z-[1] flex min-h-full flex-col">
      <GuestNav />
      <main class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-lg flex-1 flex-col justify-center px-4 py-10 sm:px-6">
        <div class="fx-auth-card">
          <h1 class="fx-auth-title">{{ $t("auth.login.title") }}</h1>
          <p class="fx-auth-subtitle">{{ $t("auth.login.subtitle") }}</p>

          <div
            v-if="err"
            id="login-error"
            role="alert"
            aria-live="assertive"
            class="fx-auth-alert"
          >
            <svg class="fx-auth-alert__icon h-5 w-5 shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
            </svg>
            <span>{{ err }}</span>
          </div>

          <form class="mt-8 space-y-5" :aria-busy="submitting" :aria-describedby="err ? 'login-error' : undefined" @submit.prevent="submit">
            <AuthFormInput
              id="username"
              v-model="username"
              icon="user"
              :label="$t('auth.login.username')"
              autocomplete="username"
              :placeholder="$t('auth.login.usernamePlaceholder')"
              :disabled="submitting"
            />
            <AuthFormInput
              id="password"
              v-model="password"
              password
              :label="$t('auth.login.password')"
              autocomplete="current-password"
              :placeholder="$t('auth.login.passwordMask')"
              :disabled="submitting"
            />
            <button type="submit" class="fx-auth-submit" :disabled="submitting" :aria-busy="submitting">
              <span v-if="submitting" class="inline-flex items-center justify-center gap-2.5">
                <svg class="fx-auth-submit-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" aria-hidden="true">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {{ $t("auth.login.submitting") }}
              </span>
              <span v-else>{{ $t("auth.login.submit") }}</span>
            </button>
          </form>

          <div class="fx-auth-card-divider" role="presentation" />
          <p class="fx-auth-footer">
            {{ $t("auth.login.footerPrompt") }}
            <RouterLink to="/register" class="fx-auth-footer-link">{{ $t("auth.login.footerLink") }}</RouterLink>
          </p>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import AuthBackground from "../components/AuthBackground.vue";
import AuthFormInput from "../components/AuthFormInput.vue";
import GuestNav from "../components/GuestNav.vue";
import { postJson } from "../api";
import { useSession } from "../session";
import FxAlert from "../components/primitives/FxAlert.vue";
import FxButton from "../components/primitives/FxButton.vue";

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

          <FxAlert v-if="err" id="login-error" size="lg" class="mt-4">{{ err }}</FxAlert>

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
            <FxButton type="submit" size="lg" full-width :loading="submitting" :disabled="submitting">
              {{ submitting ? $t("auth.login.submitting") : $t("auth.login.submit") }}
            </FxButton>
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

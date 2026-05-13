<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink, useRoute, useRouter } from "vue-router";
import AuthBackground from "../components/AuthBackground.vue";
import AuthFormInput from "../components/AuthFormInput.vue";
import GuestNav from "../components/GuestNav.vue";
import { api, postJson } from "../api";
import { useSession } from "../session";
import {
  registerEmailFormatOk,
  registerPasswordLengthOk,
  registerUsernameLengthOk,
} from "../utils/registerValidation";

type Bootstrap = {
  user_count: number;
  has_registration_mode: boolean;
  registration_mode: string;
};

type UsernameRemote = "idle" | "checking" | "available" | "taken" | "invalid_length" | "error";

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
const submitting = ref(false);
const attemptedSubmit = ref(false);

const usernameRemote = ref<UsernameRemote>("idle");
let usernameDebounce: ReturnType<typeof setTimeout> | undefined;

const hasInvite = computed(() => invite.value.trim().length > 0);

const help = computed(() => {
  const b = boot.value;
  if (!b) return "";
  if (b.user_count === 0) return t("auth.register.helpFirstAdmin");
  if (b.has_registration_mode && b.registration_mode === "invite") return t("auth.register.helpInvite");
  if (b.has_registration_mode && b.registration_mode === "open") return t("auth.register.helpOpen");
  return t("auth.register.helpDefault");
});

const usernameFieldError = computed(() => {
  const u = username.value;
  if (!attemptedSubmit.value && u.trim() === "") return "";
  if (u.trim() === "") return t("auth.register.validationUsernameRequired");
  if (!registerUsernameLengthOk(u)) return t("auth.register.validationUsernameLength");
  if (usernameRemote.value === "taken") return t("auth.register.validationUsernameTaken");
  return "";
});

const emailFieldError = computed(() => {
  if (!attemptedSubmit.value && email.value.trim() === "") return "";
  if (email.value.trim() === "") return t("auth.register.validationEmailInvalid");
  if (!registerEmailFormatOk(email.value)) return t("auth.register.validationEmailInvalid");
  return "";
});

const passwordFieldError = computed(() => {
  if (!attemptedSubmit.value && password.value === "") return "";
  if (!registerPasswordLengthOk(password.value)) return t("auth.register.validationPasswordLength");
  return "";
});

const showUsernameAvailable = computed(
  () =>
    registerUsernameLengthOk(username.value) &&
    usernameRemote.value === "available" &&
    !usernameFieldError.value,
);

watch(username, () => {
  clearTimeout(usernameDebounce);
  if (!registerUsernameLengthOk(username.value)) {
    usernameRemote.value = "idle";
    return;
  }
  usernameRemote.value = "checking";
  usernameDebounce = setTimeout(async () => {
    if (!registerUsernameLengthOk(username.value)) {
      usernameRemote.value = "idle";
      return;
    }
    try {
      const q = new URLSearchParams({ username: username.value.trim() });
      const r = await api<{ available: boolean; reason?: string }>(`/api/auth/username-available?${q.toString()}`);
      if (r.available) {
        usernameRemote.value = "available";
      } else if (r.reason === "taken") {
        usernameRemote.value = "taken";
      } else {
        usernameRemote.value = "invalid_length";
      }
    } catch {
      usernameRemote.value = "error";
    }
  }, 400);
});

onMounted(async () => {
  try {
    boot.value = await api<Bootstrap>("/api/bootstrap");
  } catch {
    boot.value = { user_count: 0, has_registration_mode: false, registration_mode: "" };
  }
});

async function checkUsernameNow(): Promise<boolean> {
  if (!registerUsernameLengthOk(username.value)) return false;
  try {
    const q = new URLSearchParams({ username: username.value.trim() });
    const r = await api<{ available: boolean; reason?: string }>(`/api/auth/username-available?${q.toString()}`);
    if (r.available) {
      usernameRemote.value = "available";
      return true;
    }
    usernameRemote.value = r.reason === "taken" ? "taken" : "invalid_length";
    return false;
  } catch {
    usernameRemote.value = "error";
    return false;
  }
}

async function submit() {
  if (submitting.value) return;
  err.value = "";
  attemptedSubmit.value = true;

  if (!registerUsernameLengthOk(username.value)) return;
  if (!registerEmailFormatOk(email.value)) return;
  if (!registerPasswordLengthOk(password.value)) return;

  const nameFree = await checkUsernameNow();
  if (!nameFree) return;

  submitting.value = true;
  try {
    const r = await postJson<{ next: string }>("/api/auth/register", {
      username: username.value.trim(),
      email: email.value.trim(),
      password: password.value,
      invite: invite.value,
    });
    await refresh();
    await router.push(r.next || "/");
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.registrationFailed");
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
          <h1 class="fx-auth-title">{{ $t("auth.register.title") }}</h1>
          <div v-if="!boot" class="fx-auth-help-skeleton" aria-hidden="true" />
          <p v-else class="fx-auth-subtitle">{{ help }}</p>

          <div v-if="hasInvite" class="fx-auth-invite-banner" role="status">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
            </svg>
            <span>{{ $t("auth.register.inviteNotice") }}</span>
          </div>

          <div
            v-if="err"
            id="register-error"
            role="alert"
            aria-live="assertive"
            class="fx-auth-alert"
          >
            <svg class="fx-auth-alert__icon h-5 w-5 shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
            </svg>
            <span>{{ err }}</span>
          </div>

          <form
            class="space-y-5"
            :class="err || hasInvite || !boot ? 'mt-6' : 'mt-8'"
            :aria-busy="submitting"
            :aria-describedby="err ? 'register-error' : undefined"
            @submit.prevent="submit"
          >
            <div>
              <AuthFormInput
                id="reg-user"
                v-model="username"
                icon="user"
                :label="$t('auth.register.username')"
                autocomplete="username"
                :placeholder="$t('auth.register.usernamePlaceholder')"
                :disabled="submitting"
                :error-message="usernameFieldError"
              />
              <p v-if="showUsernameAvailable" class="fx-auth-username-available">{{ $t("auth.register.usernameAvailable") }}</p>
              <p v-else-if="usernameRemote === 'checking' && registerUsernameLengthOk(username)" class="fx-auth-username-checking">
                {{ $t("auth.register.checkingUsername") }}
              </p>
            </div>
            <AuthFormInput
              id="reg-email"
              v-model="email"
              icon="email"
              :label="$t('auth.register.email')"
              autocomplete="email"
              :placeholder="$t('auth.register.emailPlaceholder')"
              :disabled="submitting"
              :error-message="emailFieldError"
            />
            <div>
              <AuthFormInput
                id="reg-pass"
                v-model="password"
                password
                :label="$t('auth.register.password')"
                autocomplete="new-password"
                :placeholder="$t('auth.register.passwordPlaceholder')"
                :disabled="submitting"
                :error-message="passwordFieldError"
              />
              <p v-if="!passwordFieldError" class="fx-auth-password-hint">{{ $t("auth.register.passwordHint") }}</p>
            </div>
            <button type="submit" class="fx-auth-submit" :disabled="submitting" :aria-busy="submitting">
              <span v-if="submitting" class="inline-flex items-center justify-center gap-2.5">
                <svg class="fx-auth-submit-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" aria-hidden="true">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {{ $t("auth.register.submitting") }}
              </span>
              <span v-else>{{ $t("auth.register.submit") }}</span>
            </button>
          </form>

          <div class="fx-auth-card-divider" role="presentation" />
          <p class="fx-auth-footer">
            {{ $t("auth.register.footerPrompt") }}
            <RouterLink to="/login" class="fx-auth-footer-link">{{ $t("auth.register.footerLink") }}</RouterLink>
          </p>
        </div>
      </main>
    </div>
  </div>
</template>

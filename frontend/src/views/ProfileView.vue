<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { api, patchJson, postJson } from "../api";
import type { User } from "../api";
import { useI18n } from "vue-i18n";
import { toast } from "../composables/useToast";
import { useSession } from "../session";
import { FX_THEMES, normalizeFxThemeId } from "../themes";
import {
  bcp47ForUiLocale,
  isSupportedLocale,
  localeMenuLabel,
  SUPPORTED_LOCALES,
  type SupportedLocale,
} from "../locale/constants";
import { usernameInitial } from "../utils/initial";
import FxSvg from "../components/FxSvg.vue";

const u = ref<User | null>(null);
const username = ref("");
const email = ref("");
const currentPassword = ref("");
const newPassword = ref("");
const removeAvatar = ref(false);
const avatar = ref<File | null>(null);
const avatarPreviewUrl = ref<string | null>(null);
const err = ref("");
const saving = ref(false);
const themeSaving = ref<string | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);
const { refresh } = useSession();
const { t, locale } = useI18n();

onMounted(async () => {
  const r = await api<{ user: User }>("/api/profile");
  u.value = r.user;
  username.value = r.user.username;
  email.value = r.user.email;
});

onUnmounted(() => {
  if (avatarPreviewUrl.value) URL.revokeObjectURL(avatarPreviewUrl.value);
});

watch(avatar, (f) => {
  if (avatarPreviewUrl.value) {
    URL.revokeObjectURL(avatarPreviewUrl.value);
    avatarPreviewUrl.value = null;
  }
  if (f) avatarPreviewUrl.value = URL.createObjectURL(f);
});

watch(removeAvatar, (rm) => {
  if (rm && avatar.value) {
    avatar.value = null;
  }
});

const serverPhotoSrc = computed(() => {
  if (!u.value?.avatar_path) return "";
  const q = encodeURIComponent(u.value.updated_at || u.value.id);
  return "/profile/photo?t=" + q;
});

const avatarImgSrc = computed(() => {
  if (avatarPreviewUrl.value) return avatarPreviewUrl.value;
  if (removeAvatar.value) return "";
  if (u.value?.avatar_path) return serverPhotoSrc.value;
  return "";
});

const roleLabel = computed(() => (u.value?.role === "admin" ? t("role.administrator") : t("role.member")));

const localeSelectOptions = computed(() =>
  SUPPORTED_LOCALES.map((code) => ({
    code,
    label: localeMenuLabel(code),
  })),
);

const activeThemeId = computed(() => normalizeFxThemeId(u.value?.theme));

async function selectTheme(themeId: string) {
  if (!u.value || themeId === activeThemeId.value) return;
  themeSaving.value = themeId;
  try {
    const r = await patchJson<{ user: User }>("/api/profile/theme", { theme: themeId });
    u.value = r.user;
    await refresh();
    toast.success(t("profile.themeUpdated"));
  } catch (e) {
    toast.error(e instanceof Error ? e.message : t("profile.themeUpdateFailed"));
  } finally {
    themeSaving.value = null;
  }
}

function formatProfileDate(iso: string | undefined): string {
  if (!iso) return t("common.emDash");
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return t("common.emDash");
  return d.toLocaleDateString(bcp47ForUiLocale(locale.value as SupportedLocale), {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function onLocaleChange(e: Event) {
  const v = (e.target as HTMLSelectElement).value;
  if (isSupportedLocale(v)) {
    locale.value = v;
  }
}

function pickImageFile(f: File | undefined) {
  if (!f || !f.type.startsWith("image/")) return;
  avatar.value = f;
  removeAvatar.value = false;
}

function onAv(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0];
  if (f) pickImageFile(f);
  else avatar.value = null;
}

function onPhotoDrop(e: DragEvent) {
  pickImageFile(e.dataTransfer?.files?.[0]);
}

function triggerFilePick() {
  fileInputRef.value?.click();
}

function clearPendingAvatar() {
  avatar.value = null;
  if (fileInputRef.value) fileInputRef.value.value = "";
}

async function save() {
  err.value = "";
  saving.value = true;
  try {
    if (avatar.value) {
      const fd = new FormData();
      fd.append("username", username.value);
      fd.append("email", email.value);
      fd.append("current_password", currentPassword.value);
      fd.append("new_password", newPassword.value);
      if (removeAvatar.value) fd.append("remove_avatar", "on");
      fd.append("avatar", avatar.value);
      await api("/api/profile", { method: "POST", body: fd });
    } else {
      await postJson("/api/profile", {
        username: username.value,
        email: email.value,
        current_password: currentPassword.value,
        new_password: newPassword.value,
        remove_avatar: removeAvatar.value,
      });
    }
    await refresh();
    const r = await api<{ user: User }>("/api/profile");
    u.value = r.user;
    currentPassword.value = "";
    newPassword.value = "";
    removeAvatar.value = false;
    avatar.value = null;
    if (fileInputRef.value) fileInputRef.value.value = "";
    toast.success(t("profile.updated"));
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.couldNotSave");
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div v-if="!u" class="mx-auto max-w-4xl text-zinc-500">{{ $t("common.loading") }}</div>
  <div v-else class="mx-auto max-w-4xl space-y-8">
    <header class="mb-2">
      <h1 class="text-2xl font-semibold tracking-tight text-zinc-900 sm:text-3xl">{{ $t("profile.title") }}</h1>
      <p class="mt-2 max-w-2xl text-zinc-500">{{ $t("profile.subtitle") }}</p>
    </header>

    <p
      v-if="err"
      class="rounded-2xl border border-red-200/90 bg-red-50 px-4 py-3 text-sm text-red-800 shadow-sm ring-1 ring-red-900/5"
      role="alert"
    >
      {{ err }}
    </p>

    <form class="space-y-6" @submit.prevent="save">
      <!-- Identity hero -->
      <section
        class="fx-card relative overflow-hidden p-0 shadow-md ring-zinc-950/[0.04]"
        aria-labelledby="profile-identity-heading"
      >
        <div
          class="pointer-events-none absolute inset-0 bg-gradient-to-br from-sky-50/90 via-white to-indigo-50/40"
          aria-hidden="true"
        ></div>
        <div class="relative flex flex-col gap-6 p-6 sm:flex-row sm:items-start sm:gap-8 sm:p-8">
          <div class="mx-auto flex w-28 shrink-0 flex-col sm:mx-0 sm:w-32">
            <input
              id="pav"
              ref="fileInputRef"
              type="file"
              accept="image/*"
              class="sr-only"
              @change="onAv"
            />
            <div
              class="relative aspect-square w-full shrink-0 overflow-hidden rounded-3xl bg-white shadow-lg shadow-zinc-900/10 ring-2 ring-white"
              @dragover.prevent
              @drop.prevent="onPhotoDrop"
            >
              <img
                v-if="avatarImgSrc"
                :src="avatarImgSrc"
                alt=""
                class="h-full w-full object-cover"
                width="128"
                height="128"
              />
              <div
                v-else
                class="flex h-full w-full items-center justify-center bg-gradient-to-br from-sky-500 to-sky-700 text-3xl font-bold text-white sm:text-4xl"
              >
                {{ usernameInitial(username) }}
              </div>
              <span
                v-if="removeAvatar && u.avatar_path && !avatarPreviewUrl"
                class="absolute inset-0 flex items-center justify-center bg-zinc-900/55 text-center text-xs font-semibold uppercase tracking-wide text-white backdrop-blur-[2px]"
              >
                {{ $t("profile.removedOnSave") }}
              </span>
            </div>
            <div class="mt-4 flex w-full flex-col gap-2">
              <button
                type="button"
                class="fx-btn-secondary w-full justify-center py-2 text-xs font-semibold"
                @click="triggerFilePick"
              >
                {{ $t("profile.changePhoto") }}
              </button>
              <button
                v-if="avatar"
                type="button"
                class="fx-btn-secondary w-full justify-center border-dashed py-2 text-xs font-medium text-zinc-600"
                @click="clearPendingAvatar"
              >
                {{ $t("profile.clearNewPhoto") }}
              </button>
              <label
                v-if="u.avatar_path || avatarPreviewUrl"
                class="flex w-full cursor-pointer items-start gap-2 rounded-xl border border-zinc-200/80 bg-white/70 px-2 py-2 text-left shadow-sm ring-1 ring-zinc-950/[0.02] transition hover:bg-white"
              >
                <input v-model="removeAvatar" type="checkbox" class="mt-0.5 shrink-0 rounded border-zinc-300 text-sky-600 focus:ring-sky-500" />
                <span class="min-w-0 text-xs leading-snug text-zinc-600">
                  <span class="font-medium text-zinc-800">{{ $t("profile.removePhoto") }}</span>
                  <span class="mt-0.5 block text-[11px] text-zinc-500">{{ $t("profile.removePhotoHelp") }}</span>
                </span>
              </label>
            </div>
          </div>
          <div class="min-w-0 flex-1 text-center sm:text-left">
            <p id="profile-identity-heading" class="text-xl font-semibold tracking-tight text-zinc-900 sm:text-2xl">
              {{ username }}
            </p>
            <p class="mt-1 truncate text-sm text-zinc-500">{{ email }}</p>
            <div class="mt-4 flex flex-wrap items-center justify-center gap-2 sm:justify-start">
              <span
                class="inline-flex items-center rounded-full bg-white/90 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-sky-800 shadow-sm ring-1 ring-sky-200/80"
                >{{ roleLabel }}</span
              >
              <span
                v-if="u.is_active"
                class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-800 ring-1 ring-emerald-200/80"
              >
                <span class="h-1.5 w-1.5 rounded-full bg-emerald-500" aria-hidden="true"></span>
                {{ $t("common.active") }}
              </span>
              <span
                v-else
                class="inline-flex items-center rounded-full bg-zinc-100 px-3 py-1 text-xs font-semibold text-zinc-600 ring-1 ring-zinc-200/80"
                >{{ $t("common.inactive") }}</span
              >
            </div>
            <dl class="mx-auto mt-6 max-w-md text-left text-sm sm:mx-0">
              <div class="rounded-xl border border-zinc-200/60 bg-white/60 px-3 py-2.5 backdrop-blur-sm">
                <dt class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("profile.memberSince") }}</dt>
                <dd class="mt-0.5 text-zinc-800">{{ formatProfileDate(u.created_at) }}</dd>
              </div>
            </dl>
          </div>
        </div>
      </section>

      <!-- Account -->
      <section class="fx-card overflow-hidden p-0" aria-labelledby="profile-account-heading">
        <div class="flex items-center gap-3 border-b border-zinc-100 px-5 py-4">
          <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="pencilSquare" class="fx-icon" /></span>
          <div>
            <h2 id="profile-account-heading" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("profile.account") }}</h2>
            <p class="text-xs text-zinc-400">{{ $t("profile.accountHelp") }}</p>
          </div>
        </div>
        <div class="space-y-4 p-5 sm:p-6">
          <div>
            <label class="fx-label" for="pu">{{ $t("auth.login.username") }}</label>
            <input id="pu" v-model="username" class="fx-input" required autocomplete="username" />
          </div>
          <div>
            <label class="fx-label" for="pe">{{ $t("auth.register.email") }}</label>
            <input id="pe" v-model="email" type="email" class="fx-input" required autocomplete="email" />
          </div>
        </div>
      </section>

      <!-- Language -->
      <section class="fx-card overflow-hidden p-0" aria-labelledby="profile-language-heading">
        <div class="flex items-center gap-3 border-b border-zinc-100 px-5 py-4">
          <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="gear" class="fx-icon" /></span>
          <div>
            <h2 id="profile-language-heading" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("profile.language") }}</h2>
            <p class="text-xs text-zinc-400">{{ $t("profile.languageHelp") }}</p>
          </div>
        </div>
        <div class="p-5 sm:p-6">
          <label class="fx-label" for="profile-locale">{{ $t("profile.language") }}</label>
          <select
            id="profile-locale"
            class="fx-input mt-1.5 max-w-xs"
            :value="locale"
            @change="onLocaleChange"
          >
            <option v-for="opt in localeSelectOptions" :key="opt.code" :value="opt.code">{{ opt.label }}</option>
          </select>
        </div>
      </section>

      <!-- Appearance (theme saves immediately; no password) -->
      <section class="fx-card overflow-hidden p-0" aria-labelledby="profile-theme-heading">
        <div class="flex items-center gap-3 border-b border-zinc-100 px-5 py-4">
          <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="eye" class="fx-icon" /></span>
          <div>
            <h2 id="profile-theme-heading" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("profile.appearance") }}</h2>
            <p class="text-xs text-zinc-400">{{ $t("profile.appearanceHelp") }}</p>
          </div>
        </div>
        <div class="grid gap-3 p-5 sm:grid-cols-2 lg:grid-cols-3 sm:p-6">
          <button
            v-for="theme in FX_THEMES"
            :key="theme.id"
            type="button"
            class="group fx-profile-theme-tile"
            :class="{ 'fx-profile-theme-tile--selected': activeThemeId === theme.id }"
            :disabled="themeSaving !== null"
            :aria-pressed="activeThemeId === theme.id"
            :aria-busy="themeSaving === theme.id"
            @click="selectTheme(theme.id)"
          >
            <div class="flex h-12 w-full gap-0.5 overflow-hidden rounded-t-xl px-2 pt-2">
              <div v-for="(sw, i) in theme.swatches" :key="i" class="min-h-0 flex-1 rounded-md shadow-inner" :class="sw" />
            </div>
            <div class="flex flex-1 flex-col gap-0.5 px-3 pb-3 pt-2.5">
              <span class="text-sm font-semibold text-zinc-900">{{ $t("themes." + theme.id + ".label") }}</span>
              <span class="text-xs leading-snug text-zinc-500">{{ $t("themes." + theme.id + ".description") }}</span>
            </div>
            <span
              v-if="themeSaving === theme.id"
              class="flex items-center justify-center gap-2 border-t border-zinc-100/80 bg-zinc-50/90 py-2 text-xs font-medium text-zinc-600"
            >
              <span
                class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-600"
                aria-hidden="true"
              ></span>
              {{ $t("profile.applying") }}
            </span>
            <span
              v-else-if="activeThemeId === theme.id"
              class="fx-profile-theme-tile__footer-active border-t py-1.5 text-center text-[11px] font-semibold uppercase tracking-wide"
            >
              {{ $t("common.active") }}
            </span>
          </button>
        </div>
      </section>

      <!-- Security -->
      <section class="fx-card overflow-hidden p-0" aria-labelledby="profile-security-heading">
        <div class="flex items-center gap-3 border-b border-zinc-100 px-5 py-4">
          <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="gear" class="fx-icon" /></span>
          <div>
            <h2 id="profile-security-heading" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("profile.security") }}</h2>
            <p class="text-xs text-zinc-400">{{ $t("profile.securityHelp") }}</p>
          </div>
        </div>
        <div class="space-y-4 p-5 sm:p-6">
          <div>
            <label class="fx-label" for="pc">{{ $t("profile.currentPassword") }}</label>
            <input id="pc" v-model="currentPassword" type="password" class="fx-input" autocomplete="current-password" />
          </div>
          <div>
            <label class="fx-label" for="pn">{{ $t("profile.newPassword") }}</label>
            <input id="pn" v-model="newPassword" type="password" class="fx-input" autocomplete="new-password" />
            <p class="mt-1.5 text-xs text-zinc-500">{{ $t("profile.newPasswordHelp") }}</p>
          </div>
        </div>
      </section>

      <div class="flex flex-col items-stretch gap-3 border-t border-zinc-200/80 pt-2 sm:flex-row sm:items-center sm:justify-end">
        <button type="submit" class="fx-btn-primary min-w-[10rem] sm:min-w-[11rem]" :disabled="saving">
          <span v-if="saving" class="inline-flex items-center gap-2">
            <span
              class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"
              aria-hidden="true"
            ></span>
            {{ $t("profile.saving") }}
          </span>
          <span v-else>{{ $t("profile.saveChanges") }}</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";
import FxToggle from "../components/primitives/FxToggle.vue";
import FxAlert from "../components/primitives/FxAlert.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";

const { t } = useI18n();

type UserRow = {
  id: string;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
};
type Invite = {
  ID: string;
  Token: string;
  Role: string;
  ExpiresAt: string;
  UsedAt?: string | null;
};

const users = ref<UserRow[]>([]);
const invites = ref<Invite[]>([]);
const err = ref("");

const newUser = ref({ username: "", email: "", password: "", role: "user" });
const newInvite = ref({ role: "user", ttl_hours: 72 });

onMounted(load);

async function load() {
  err.value = "";
  try {
    const j = await api<{ users: UserRow[]; invites: Invite[] }>("/api/admin/users");
    users.value = j.users;
    invites.value = j.invites;
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.loadFailed");
  }
}

async function setRole(uid: string, role: string) {
  err.value = "";
  try {
    await postJson("/api/admin/users/" + uid + "/role", { role });
    await load();
    toast.success(t("toast.roleUpdated"));
  } catch (e) {
    const msg = e instanceof Error ? e.message : t("toast.updateFailed");
    err.value = msg;
    toast.error(msg);
  }
}

async function setActive(uid: string, active: boolean) {
  err.value = "";
  try {
    await postJson("/api/admin/users/" + uid + "/active", { active });
    await load();
    toast.success(active ? t("toast.userActivated") : t("toast.userDeactivated"));
  } catch (e) {
    const msg = e instanceof Error ? e.message : t("toast.updateFailed");
    err.value = msg;
    toast.error(msg);
  }
}

async function createUser() {
  err.value = "";
  try {
    await postJson("/api/admin/users", {
      username: newUser.value.username,
      email: newUser.value.email,
      password: newUser.value.password,
      role: newUser.value.role,
    });
    newUser.value = { username: "", email: "", password: "", role: "user" };
    await load();
    toast.success(t("toast.userCreated"));
  } catch (e) {
    const msg = e instanceof Error ? e.message : t("toast.createFailed");
    err.value = msg;
    toast.error(msg);
  }
}

async function createInvite() {
  err.value = "";
  try {
    await postJson("/api/admin/invites", {
      role: newInvite.value.role,
      ttl_hours: newInvite.value.ttl_hours,
    });
    await load();
    toast.success(t("toast.inviteCreated"));
  } catch (e) {
    const msg = e instanceof Error ? e.message : t("toast.createFailed");
    err.value = msg;
    toast.error(msg);
  }
}
</script>

<template>
  <div class="w-full space-y-6">
    <FxPageHeader :title="$t('adminUsers.pageTitle')">
      <template #actions>
        <FxButton variant="secondary" size="sm" icon-left="arrowDownTray" :href="'/admin/backup.zip'">{{ $t("common.downloadBackup") }}</FxButton>
      </template>
    </FxPageHeader>
    <FxAlert v-if="err">{{ err }}</FxAlert>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">{{ $t("adminUsers.createUserHeading") }}</h2>
      <form class="mt-4 grid gap-3 sm:grid-cols-2" @submit.prevent="createUser">
        <input v-model="newUser.username" class="fx-input" :placeholder="$t('auth.login.username')" required />
        <input v-model="newUser.email" class="fx-input" type="email" :placeholder="$t('auth.register.email')" required />
        <input
          v-model="newUser.password"
          class="fx-input"
          type="password"
          :placeholder="$t('adminUsers.passPlaceholder')"
          required
        />
        <select v-model="newUser.role" class="fx-input">
          <option value="user">{{ $t("role.member") }}</option>
          <option value="admin">{{ $t("role.administrator") }}</option>
        </select>
        <button type="submit" class="fx-btn-primary sm:col-span-2">{{ $t("common.create") }}</button>
      </form>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">{{ $t("adminUsers.invitesHeading") }}</h2>
      <form class="mt-4 flex flex-wrap items-end gap-3" @submit.prevent="createInvite">
        <select v-model="newInvite.role" class="fx-input">
          <option value="user">{{ $t("role.member") }}</option>
          <option value="admin">{{ $t("role.administrator") }}</option>
        </select>
        <input v-model.number="newInvite.ttl_hours" type="number" min="1" class="fx-input w-32" />
        <span class="text-sm text-zinc-500">{{ $t("adminUsers.inviteHoursLabel") }}</span>
        <button type="submit" class="fx-btn-secondary text-sm">{{ $t("adminUsers.createInvite") }}</button>
      </form>
      <ul class="mt-4 divide-y divide-zinc-100">
        <li v-for="inv in invites" :key="inv.ID" class="py-3 font-mono text-xs break-all">
          <span class="font-sans text-sm text-zinc-700">{{
            inv.Role === "admin" ? $t("role.administrator") : $t("role.member")
          }}</span>
          · {{ $t("common.expiresWord") }} {{ inv.ExpiresAt }}
          <RouterLink v-if="!inv.UsedAt" class="ml-2 text-sky-700" :to="'/register?invite=' + encodeURIComponent(inv.Token)">{{
            $t("common.openLink")
          }}</RouterLink>
        </li>
      </ul>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">{{ $t("adminUsers.usersHeading") }}</h2>
      <table class="mt-4 w-full text-left text-sm">
        <thead>
          <tr class="border-b border-zinc-200 text-zinc-500">
            <th class="py-2 pr-2">{{ $t("adminUsers.colUser") }}</th>
            <th class="py-2 pr-2">{{ $t("adminUsers.colRole") }}</th>
            <th class="py-2 pr-2">{{ $t("adminUsers.colActive") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-b border-zinc-100">
            <td class="py-2 pr-2">
              <div class="font-medium">{{ u.username }}</div>
              <div class="text-xs text-zinc-500">{{ u.email }}</div>
            </td>
            <td class="py-2 pr-2">
              <select :value="u.role" class="fx-input py-1 text-xs" @change="setRole(u.id, ($event.target as HTMLSelectElement).value)">
                <option value="user">{{ $t("role.member") }}</option>
                <option value="admin">{{ $t("role.administrator") }}</option>
              </select>
            </td>
            <td class="py-2 pr-2">
              <FxToggle
                :model-value="u.is_active"
                :aria-label="u.is_active ? $t('common.active') : $t('common.inactive')"
                @update:model-value="setActive(u.id, $event)"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

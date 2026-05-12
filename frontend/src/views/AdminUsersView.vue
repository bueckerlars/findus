<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";

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
    err.value = e instanceof Error ? e.message : "Load failed";
  }
}

async function setRole(uid: string, role: string) {
  err.value = "";
  try {
    await postJson("/api/admin/users/" + uid + "/role", { role });
    await load();
    toast.success("Role updated.");
  } catch (e) {
    const msg = e instanceof Error ? e.message : "Update failed";
    err.value = msg;
    toast.error(msg);
  }
}

async function setActive(uid: string, active: boolean) {
  err.value = "";
  try {
    await postJson("/api/admin/users/" + uid + "/active", { active });
    await load();
    toast.success(active ? "User activated." : "User deactivated.");
  } catch (e) {
    const msg = e instanceof Error ? e.message : "Update failed";
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
    toast.success("User created.");
  } catch (e) {
    const msg = e instanceof Error ? e.message : "Create failed";
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
    toast.success("Invite created.");
  } catch (e) {
    const msg = e instanceof Error ? e.message : "Create failed";
    err.value = msg;
    toast.error(msg);
  }
}
</script>

<template>
  <div class="max-w-4xl space-y-10">
    <div class="flex flex-wrap items-center justify-between gap-4">
      <h1 class="text-2xl font-semibold text-zinc-900">User management</h1>
      <a href="/admin/backup.zip" class="fx-btn-secondary text-sm">Download backup</a>
    </div>
    <p v-if="err" class="text-sm text-red-700">{{ err }}</p>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">Create user</h2>
      <form class="mt-4 grid gap-3 sm:grid-cols-2" @submit.prevent="createUser">
        <input v-model="newUser.username" class="fx-input" placeholder="Username" required />
        <input v-model="newUser.email" class="fx-input" type="email" placeholder="Email" required />
        <input v-model="newUser.password" class="fx-input" type="password" placeholder="Password" required />
        <select v-model="newUser.role" class="fx-input">
          <option value="user">User</option>
          <option value="admin">Admin</option>
        </select>
        <button type="submit" class="fx-btn-primary sm:col-span-2">Create</button>
      </form>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">Invites</h2>
      <form class="mt-4 flex flex-wrap items-end gap-3" @submit.prevent="createInvite">
        <select v-model="newInvite.role" class="fx-input">
          <option value="user">User</option>
          <option value="admin">Admin</option>
        </select>
        <input v-model.number="newInvite.ttl_hours" type="number" min="1" class="fx-input w-32" />
        <span class="text-sm text-zinc-500">hours</span>
        <button type="submit" class="fx-btn-secondary text-sm">Create invite</button>
      </form>
      <ul class="mt-4 divide-y divide-zinc-100">
        <li v-for="inv in invites" :key="inv.ID" class="py-3 font-mono text-xs break-all">
          <span class="font-sans text-sm text-zinc-700">{{ inv.Role }}</span>
          · expires {{ inv.ExpiresAt }}
          <RouterLink v-if="!inv.UsedAt" class="ml-2 text-sky-700" :to="'/register?invite=' + encodeURIComponent(inv.Token)">Open link</RouterLink>
        </li>
      </ul>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">Users</h2>
      <table class="mt-4 w-full text-left text-sm">
        <thead>
          <tr class="border-b border-zinc-200 text-zinc-500">
            <th class="py-2 pr-2">User</th>
            <th class="py-2 pr-2">Role</th>
            <th class="py-2 pr-2">Active</th>
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
                <option value="user">user</option>
                <option value="admin">admin</option>
              </select>
            </td>
            <td class="py-2 pr-2">
              <input type="checkbox" :checked="u.is_active" @change="setActive(u.id, ($event.target as HTMLInputElement).checked)" />
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

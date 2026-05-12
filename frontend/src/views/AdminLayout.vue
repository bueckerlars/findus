<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

const route = useRoute();
const { t } = useI18n();
const usersActive = computed(() => route.path === "/admin/users");
const settingsActive = computed(() => route.path === "/admin/settings");
const templatesActive = computed(() => route.path.startsWith("/admin/templates"));

function tabClass(active: boolean) {
  return active
    ? "border-zinc-200/90 border-b-white bg-white text-zinc-900 shadow-sm ring-1 ring-zinc-950/[0.04]"
    : "border-transparent";
}
</script>

<template>
  <div class="max-w-4xl space-y-8">
    <div class="border-b border-zinc-200/90">
      <nav class="-mb-px flex flex-wrap gap-1" :aria-label="t('adminLayout.navAria')">
        <RouterLink
          to="/admin/users"
          class="inline-flex items-center rounded-t-lg border px-4 py-2.5 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
          :class="tabClass(usersActive)"
        >
          {{ t("adminLayout.tabUsers") }}
        </RouterLink>
        <RouterLink
          to="/admin/settings"
          class="inline-flex items-center rounded-t-lg border px-4 py-2.5 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
          :class="tabClass(settingsActive)"
        >
          {{ t("adminLayout.tabSettings") }}
        </RouterLink>
        <RouterLink
          to="/admin/templates"
          class="inline-flex items-center rounded-t-lg border px-4 py-2.5 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
          :class="tabClass(templatesActive)"
        >
          {{ t("adminLayout.tabTemplates") }}
        </RouterLink>
      </nav>
    </div>
    <RouterView />
  </div>
</template>

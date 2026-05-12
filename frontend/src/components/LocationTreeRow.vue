<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";
import FxSvg from "./FxSvg.vue";
import FxQrMenuButton from "./FxQrMenuButton.vue";
import LocationTreeRow from "./LocationTreeRow.vue";

const { t } = useI18n();

export type LocationTreeNode = { ID: string; Name: string; children: LocationTreeNode[] };

const props = defineProps<{
  node: LocationTreeNode;
  isExpanded: (id: string) => boolean;
  toggle: (id: string) => void;
  isAdmin: boolean;
}>();

const hasKids = computed(() => props.node.children.length > 0);
const open = computed(() => props.isExpanded(props.node.ID));

function onExpandClick() {
  if (!hasKids.value) return;
  props.toggle(props.node.ID);
}
</script>

<template>
  <li class="relative list-none">
    <div
      class="flex min-h-[3.25rem] items-center gap-1.5 px-3 py-2.5 transition-colors sm:gap-2 sm:px-4 sm:py-3"
      :class="hasKids ? 'hover:bg-zinc-50/90' : 'hover:bg-zinc-50/80'"
    >
      <button
        v-if="hasKids"
        type="button"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-zinc-500 outline-offset-2 transition hover:bg-zinc-100 hover:text-zinc-800 focus-visible:outline focus-visible:ring-2 focus-visible:ring-sky-300/50"
        :aria-expanded="open ? 'true' : 'false'"
        :aria-label="open ? t('locationTreeRow.collapse') : t('locationTreeRow.expand')"
        @click="onExpandClick"
      >
        <FxSvg
          name="chevronRight"
          class="h-4 w-4 shrink-0 text-current transition-transform duration-200 ease-out motion-reduce:transition-none"
          :class="open ? 'rotate-90' : ''"
        />
      </button>
      <span v-else class="h-9 w-9 shrink-0" aria-hidden="true"></span>

      <RouterLink
        :to="'/locations/' + node.ID"
        class="min-w-0 flex-1 truncate text-left text-[15px] font-medium text-zinc-900 outline-offset-2 transition-colors hover:text-sky-800"
      >
        {{ node.Name }}
      </RouterLink>

      <div class="flex shrink-0 items-center gap-1" @click.stop>
        <FxQrMenuButton
          :png-url="'/locations/' + node.ID + '/qr.png'"
          :download-name="node.Name"
          :hint="t('locationDetail.qrHint')"
        />
        <RouterLink
          v-if="isAdmin"
          :to="'/locations/' + node.ID + '/edit'"
          class="fx-icon-btn"
          :aria-label="t('locationDetail.editLocation')"
          :title="t('locationDetail.edit')"
        >
          <FxSvg name="pencilSquare" />
        </RouterLink>
      </div>
    </div>

    <ul
      v-if="hasKids && open"
      class="divide-y divide-zinc-100/90 border-t border-zinc-100 bg-zinc-50/50"
      role="group"
      :aria-label="t('locationTreeRow.subLocationsOf', { name: node.Name })"
    >
      <LocationTreeRow
        v-for="ch in node.children"
        :key="ch.ID"
        :node="ch"
        :is-expanded="isExpanded"
        :toggle="toggle"
        :is-admin="isAdmin"
      />
    </ul>
  </li>
</template>

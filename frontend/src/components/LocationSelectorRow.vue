<script setup lang="ts">
import { computed } from "vue";
import FxSvg from "./FxSvg.vue";
import FxToggle from "./primitives/FxToggle.vue";
import LocationSelectorRow from "./LocationSelectorRow.vue";

export type LocationTreeNode = { ID: string; Name: string; children: LocationTreeNode[] };

const props = defineProps<{
  node: LocationTreeNode;
  selected: Set<string>;
  expanded: Set<string>;
}>();

const emit = defineEmits<{
  (e: "toggle-select", id: string): void;
  (e: "toggle-expand", id: string): void;
}>();

const hasKids = computed(() => props.node.children.length > 0);
const isOpen = computed(() => props.expanded.has(props.node.ID));
const isChecked = computed(() => props.selected.has(props.node.ID));
</script>

<template>
  <li class="relative list-none">
    <label
      class="flex min-h-[3.25rem] cursor-pointer items-center gap-1.5 px-3 py-2.5 transition-colors sm:gap-2 sm:px-4 sm:py-3"
      :class="hasKids ? 'hover:bg-zinc-50/90' : 'hover:bg-zinc-50/80'"
    >
      <!-- expand / collapse chevron -->
      <button
        v-if="hasKids"
        type="button"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-zinc-500 outline-offset-2 transition hover:bg-zinc-100 hover:text-zinc-800 focus-visible:outline focus-visible:ring-2 focus-visible:ring-sky-300/50"
        :aria-expanded="isOpen ? 'true' : 'false'"
        @click.prevent="emit('toggle-expand', node.ID)"
      >
        <FxSvg
          name="chevronRight"
          class="h-4 w-4 shrink-0 text-current transition-transform duration-200 ease-out motion-reduce:transition-none"
          :class="isOpen ? 'rotate-90' : ''"
        />
      </button>
      <span v-else class="h-9 w-9 shrink-0" aria-hidden="true" />

      <!-- location name -->
      <span class="min-w-0 flex-1 truncate text-[15px] font-medium text-zinc-900">
        {{ node.Name }}
      </span>

      <!-- toggle -->
      <span @click.stop>
        <FxToggle :model-value="isChecked" :aria-label="node.Name" @update:model-value="emit('toggle-select', node.ID)" />
      </span>
    </label>

    <ul
      v-if="hasKids && isOpen"
      class="divide-y divide-zinc-100/90 border-t border-zinc-100 bg-zinc-50/50"
      role="group"
    >
      <LocationSelectorRow
        v-for="child in node.children"
        :key="child.ID"
        :node="child"
        :selected="selected"
        :expanded="expanded"
        @toggle-select="(id) => emit('toggle-select', id)"
        @toggle-expand="(id) => emit('toggle-expand', id)"
      />
    </ul>
  </li>
</template>

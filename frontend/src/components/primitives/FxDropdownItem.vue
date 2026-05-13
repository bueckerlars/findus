<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import type { RouteLocationRaw } from "vue-router";
import { DropdownMenuItem } from "reka-ui";
import FxSvg from "../FxSvg.vue";
import type { FxIconName } from "../FxSvg.vue";

const props = withDefaults(
  defineProps<{
    icon?: FxIconName;
    tone?: "neutral" | "danger";
    to?: RouteLocationRaw;
    disabled?: boolean;
  }>(),
  { tone: "neutral", disabled: false, icon: undefined, to: undefined },
);

const emit = defineEmits<{ select: [Event] }>();
const router = useRouter();

const classes = computed(() => [
  "fx-menu-item",
  props.tone === "danger" ? "fx-menu-item--danger" : null,
]);

function onSelect(event: Event) {
  if (props.to !== undefined) {
    void router.push(props.to);
  }
  emit("select", event);
}
</script>

<template>
  <DropdownMenuItem :class="classes" :disabled="disabled" @select="onSelect">
    <FxSvg v-if="icon" :name="icon" class="h-4 w-4 shrink-0" />
    <span class="min-w-0 flex-1 truncate"><slot /></span>
  </DropdownMenuItem>
</template>

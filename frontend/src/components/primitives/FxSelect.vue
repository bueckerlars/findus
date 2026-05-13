<script setup lang="ts">
import { computed, useAttrs } from "vue";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    modelValue?: string | number | null;
    size?: "sm" | "md" | "lg";
    invalid?: boolean;
  }>(),
  { size: "md", invalid: false, modelValue: undefined },
);

const emit = defineEmits<{ "update:modelValue": [string] }>();

const attrs = useAttrs();

const cls = computed(() => [
  "fx-input",
  props.size === "sm" ? "fx-input--sm" : props.size === "lg" ? "fx-input--lg" : null,
  props.invalid ? "fx-input--invalid" : null,
  attrs.class as string | undefined,
]);

function onChange(e: Event) {
  emit("update:modelValue", (e.target as HTMLSelectElement).value);
}
</script>

<template>
  <select
    :value="modelValue ?? ''"
    :class="cls"
    :aria-invalid="invalid || undefined"
    v-bind="$attrs"
    @change="onChange"
  >
    <slot />
  </select>
</template>

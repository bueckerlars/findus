<script setup lang="ts">
import { computed, useAttrs } from "vue";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    modelValue?: string | null;
    size?: "sm" | "md" | "lg";
    invalid?: boolean;
    minHeightClass?: string;
  }>(),
  { size: "md", invalid: false, modelValue: "", minHeightClass: "min-h-[5rem]" },
);

const emit = defineEmits<{ "update:modelValue": [string] }>();

const attrs = useAttrs();

const cls = computed(() => [
  "fx-input",
  props.minHeightClass,
  props.size === "sm" ? "fx-input--sm" : props.size === "lg" ? "fx-input--lg" : null,
  props.invalid ? "fx-input--invalid" : null,
  attrs.class as string | undefined,
]);

function onInput(e: Event) {
  emit("update:modelValue", (e.target as HTMLTextAreaElement).value);
}
</script>

<template>
  <textarea
    :value="modelValue ?? ''"
    :class="cls"
    :aria-invalid="invalid || undefined"
    v-bind="$attrs"
    @input="onInput"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { SwitchRoot, SwitchThumb } from "reka-ui";

const props = withDefaults(
  defineProps<{
    modelValue?: boolean;
    size?: "sm" | "md";
    disabled?: boolean;
    label?: string;
    description?: string;
    inputId?: string;
    ariaLabel?: string;
  }>(),
  {
    modelValue: false,
    size: "sm",
    disabled: false,
    label: undefined,
    description: undefined,
    inputId: undefined,
    ariaLabel: undefined,
  },
);

const emit = defineEmits<{ "update:modelValue": [boolean] }>();

const classes = computed(() => [
  "fx-toggle",
  props.size === "md" ? "fx-toggle--md" : null,
]);

function onUpdate(v: boolean) {
  emit("update:modelValue", v);
}
</script>

<template>
  <span class="inline-flex items-center gap-2">
    <SwitchRoot
      :id="inputId"
      :model-value="modelValue"
      :disabled="disabled"
      :class="classes"
      :aria-label="ariaLabel ?? label"
      @update:model-value="onUpdate"
    >
      <SwitchThumb class="fx-toggle__thumb" />
    </SwitchRoot>
    <span v-if="label || description" class="min-w-0 leading-tight">
      <span v-if="label" class="block text-sm font-medium text-zinc-800">{{ label }}</span>
      <span v-if="description" class="block text-xs text-zinc-500">{{ description }}</span>
    </span>
  </span>
</template>

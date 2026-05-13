<script setup lang="ts">
import { computed, useId } from "vue";

const props = withDefaults(
  defineProps<{
    label?: string;
    hint?: string;
    error?: string;
    required?: boolean;
    inputId?: string;
    srOnlyLabel?: boolean;
  }>(),
  {
    label: undefined,
    hint: undefined,
    error: undefined,
    required: false,
    inputId: undefined,
    srOnlyLabel: false,
  },
);

const generatedId = useId();
const fieldId = computed(() => props.inputId ?? generatedId);
const helpId = computed(() => `${fieldId.value}-help`);
const errorId = computed(() => `${fieldId.value}-error`);

const describedBy = computed(() => {
  const ids: string[] = [];
  if (props.error) ids.push(errorId.value);
  else if (props.hint) ids.push(helpId.value);
  return ids.length ? ids.join(" ") : undefined;
});
</script>

<template>
  <div class="fx-field">
    <label
      v-if="label"
      :for="fieldId"
      :class="['fx-label', srOnlyLabel ? 'sr-only' : null]"
    >
      {{ label }}
      <span v-if="required" class="text-red-500" aria-hidden="true">*</span>
    </label>
    <slot :id="fieldId" :described-by="describedBy" :invalid="!!error" />
    <p v-if="error" :id="errorId" class="mt-1.5 text-xs font-medium text-red-600" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="helpId" class="mt-1 text-xs text-zinc-500">{{ hint }}</p>
  </div>
</template>

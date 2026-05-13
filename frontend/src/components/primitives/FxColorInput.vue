<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    disabled?: boolean;
    inputId?: string;
    ariaLabel?: string;
  }>(),
  { modelValue: "#000000", disabled: false, inputId: undefined, ariaLabel: undefined },
);

const emit = defineEmits<{ "update:modelValue": [string] }>();

const hex = computed(() => (props.modelValue || "").toUpperCase());

function onColor(e: Event) {
  emit("update:modelValue", (e.target as HTMLInputElement).value);
}

function onHex(e: Event) {
  let v = (e.target as HTMLInputElement).value.trim();
  if (v && !v.startsWith("#")) v = `#${v}`;
  emit("update:modelValue", v);
}
</script>

<template>
  <div class="inline-flex items-center gap-2">
    <label class="relative inline-flex h-9 w-9 cursor-pointer items-center justify-center overflow-hidden rounded-lg border border-zinc-200 bg-white shadow-sm">
      <input
        :id="inputId"
        type="color"
        :value="hex"
        :disabled="disabled"
        :aria-label="ariaLabel"
        class="absolute inset-0 h-full w-full cursor-pointer appearance-none border-0 bg-transparent p-0 [&::-webkit-color-swatch-wrapper]:p-0 [&::-webkit-color-swatch]:rounded-md [&::-webkit-color-swatch]:border-0"
        @input="onColor"
      />
    </label>
    <input
      type="text"
      class="fx-input fx-input--sm w-24 font-mono uppercase"
      :value="hex"
      :disabled="disabled"
      maxlength="9"
      :aria-label="ariaLabel"
      @input="onHex"
    />
  </div>
</template>

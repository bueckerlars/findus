<script setup lang="ts">
import { useId, watch, onUnmounted, nextTick, ref } from "vue";

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    title: string;
    maxWidthClass?: string;
  }>(),
  { maxWidthClass: "max-w-lg" },
);

const emit = defineEmits<{
  "update:modelValue": [boolean];
  close: [];
}>();

const titleId = useId();
const panelRef = ref<HTMLElement | null>(null);

function cleanup() {
  document.documentElement.classList.remove("fx-dialog-open");
  document.removeEventListener("keydown", onKeydown, true);
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    close();
  }
}

function close() {
  emit("update:modelValue", false);
  emit("close");
}

function onBackdropPointerDown() {
  close();
}

function focusFirstControl() {
  const root = panelRef.value;
  if (!root) return;
  const el = root.querySelector<HTMLElement>(
    'button:not([disabled]), [href], input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
  );
  el?.focus();
}

watch(
  () => props.modelValue,
  async (open) => {
    if (!open) {
      cleanup();
      return;
    }
    document.documentElement.classList.add("fx-dialog-open");
    document.addEventListener("keydown", onKeydown, true);
    await nextTick();
    focusFirstControl();
  },
  { immediate: true },
);

onUnmounted(() => {
  cleanup();
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-[101] flex items-center justify-center p-4 sm:p-6"
      role="presentation"
    >
      <div
        class="absolute inset-0 bg-zinc-950/40 backdrop-blur-[2px]"
        aria-hidden="true"
        @pointerdown="onBackdropPointerDown"
      />
      <div
        ref="panelRef"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        class="relative z-[1] flex max-h-[min(85vh,44rem)] w-full flex-col rounded-2xl border border-zinc-200/90 bg-white shadow-xl ring-1 ring-zinc-950/[0.04]"
        :class="maxWidthClass"
        @pointerdown.stop
      >
        <div class="shrink-0 border-b border-zinc-100/90 px-5 pb-4 pt-5 sm:px-6 sm:pt-6">
          <h2 :id="titleId" class="text-lg font-semibold tracking-tight text-zinc-900">
            {{ title }}
          </h2>
        </div>
        <div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-5 py-4 sm:px-6">
          <slot />
        </div>
        <div v-if="$slots.footer" class="shrink-0 border-t border-zinc-100/90 px-5 py-4 sm:px-6">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

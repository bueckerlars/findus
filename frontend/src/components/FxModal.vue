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
const tabbableSelector = [
  'button:not([disabled])',
  '[href]',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(", ");
const inputSelector = [
  'input:not([disabled]):not([type="hidden"]):not([tabindex="-1"])',
  'select:not([disabled]):not([tabindex="-1"])',
  'textarea:not([disabled]):not([tabindex="-1"])',
].join(", ");

function cleanup() {
  document.documentElement.classList.remove("fx-dialog-open");
  document.removeEventListener("keydown", onKeydown, true);
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    close();
    return;
  }
  if (e.key === "Tab") {
    trapFocus(e);
  }
}

function close() {
  emit("update:modelValue", false);
  emit("close");
}

function onBackdropPointerDown() {
  close();
}

function isVisible(el: HTMLElement) {
  return el.offsetParent !== null || el.getClientRects().length > 0;
}

function focusableControls() {
  const root = panelRef.value;
  if (!root) return [];
  return Array.from(root.querySelectorAll<HTMLElement>(tabbableSelector)).filter(isVisible);
}

function trapFocus(e: KeyboardEvent) {
  const controls = focusableControls();
  if (!controls.length) {
    e.preventDefault();
    panelRef.value?.focus();
    return;
  }

  const first = controls[0];
  const last = controls[controls.length - 1];
  const active = document.activeElement;

  if (e.shiftKey && (active === first || !panelRef.value?.contains(active))) {
    e.preventDefault();
    last.focus();
    return;
  }

  if (!e.shiftKey && active === last) {
    e.preventDefault();
    first.focus();
  }
}

let focusRetryHandle: number | null = null;

function cancelFocusRetry() {
  if (focusRetryHandle !== null) {
    cancelAnimationFrame(focusRetryHandle);
    focusRetryHandle = null;
  }
}

function findFirstInput(): HTMLElement | null {
  const root = panelRef.value;
  if (!root) return null;
  return Array.from(root.querySelectorAll<HTMLElement>(inputSelector)).find(isVisible) ?? null;
}

function applyFocus(el: HTMLElement) {
  el.focus({ preventScroll: false });
  if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
    try {
      el.select();
    } catch {
      /* ignore selection errors on non-text inputs */
    }
  }
}

function focusFirstControl(remainingFrames = 10) {
  cancelFocusRetry();
  const root = panelRef.value;
  if (!root) {
    if (remainingFrames > 0) {
      focusRetryHandle = requestAnimationFrame(() => focusFirstControl(remainingFrames - 1));
    }
    return;
  }
  const firstInput = findFirstInput();
  if (firstInput) {
    applyFocus(firstInput);
    return;
  }
  if (remainingFrames > 0) {
    focusRetryHandle = requestAnimationFrame(() => focusFirstControl(remainingFrames - 1));
    return;
  }
  const fallback = focusableControls()[0] ?? root;
  fallback.focus();
}

watch(
  () => props.modelValue,
  async (open) => {
    if (!open) {
      cancelFocusRetry();
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
  cancelFocusRetry();
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
        tabindex="-1"
        class="relative z-[1] flex max-h-[min(90dvh,44rem)] w-full flex-col rounded-2xl border border-zinc-200/90 bg-white shadow-xl ring-1 ring-zinc-950/[0.04] sm:max-h-[min(85vh,44rem)]"
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

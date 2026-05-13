<script setup lang="ts">
import { useId, onMounted, onUnmounted, nextTick, ref } from "vue";

defineProps<{
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  variant: "danger" | "default";
}>();

const emit = defineEmits<{ confirm: []; cancel: [] }>();

const titleId = useId();
const descId = useId();
const panelRef = ref<HTMLElement | null>(null);
const cancelBtnRef = ref<HTMLButtonElement | null>(null);
const tabbableSelector = [
  'button:not([disabled])',
  '[href]',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(", ");

function onBackdropPointerDown() {
  emit("cancel");
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    emit("cancel");
    return;
  }
  if (e.key === "Tab") {
    trapFocus(e);
  }
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

onMounted(async () => {
  document.documentElement.classList.add("fx-dialog-open");
  document.addEventListener("keydown", onKeydown, true);
  await nextTick();
  cancelBtnRef.value?.focus();
});

onUnmounted(() => {
  document.documentElement.classList.remove("fx-dialog-open");
  document.removeEventListener("keydown", onKeydown, true);
});
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6" role="presentation">
      <div
        class="absolute inset-0 bg-zinc-950/40 backdrop-blur-[2px]"
        aria-hidden="true"
        @pointerdown="onBackdropPointerDown"
      />
      <div
        ref="panelRef"
        role="alertdialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-describedby="message ? descId : undefined"
        tabindex="-1"
        class="relative z-[1] w-full max-w-md overflow-y-auto overscroll-contain rounded-2xl border border-zinc-200/90 bg-white p-6 shadow-xl ring-1 ring-zinc-950/[0.04] max-sm:max-h-[min(90dvh,28rem)] sm:max-h-[min(85vh,32rem)]"
        @pointerdown.stop
      >
        <h2 :id="titleId" class="text-lg font-semibold tracking-tight text-zinc-900">
          {{ title }}
        </h2>
        <p v-if="message" :id="descId" class="mt-2 text-sm leading-relaxed text-zinc-600">
          {{ message }}
        </p>
        <div class="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            ref="cancelBtnRef"
            type="button"
            class="fx-btn-secondary w-full sm:w-auto"
            @click="emit('cancel')"
          >
            {{ cancelLabel }}
          </button>
          <button
            type="button"
            class="w-full sm:w-auto"
            :class="variant === 'danger' ? 'fx-btn-danger' : 'fx-btn-primary'"
            @click="emit('confirm')"
          >
            {{ confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

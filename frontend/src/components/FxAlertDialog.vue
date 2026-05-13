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
const cancelBtnRef = ref<HTMLButtonElement | null>(null);

function onBackdropPointerDown() {
  emit("cancel");
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    emit("cancel");
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
        role="alertdialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-describedby="message ? descId : undefined"
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

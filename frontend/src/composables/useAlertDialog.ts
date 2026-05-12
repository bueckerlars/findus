import { ref, shallowRef } from "vue";
import { i18n } from "../i18n";

export type AlertDialogVariant = "danger" | "default";

export type AlertDialogOptions = {
  title: string;
  message?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: AlertDialogVariant;
};

type DialogState = {
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  variant: AlertDialogVariant;
};

const visible = ref(false);
const dialogState = shallowRef<DialogState | null>(null);
let pendingResolve: ((ok: boolean) => void) | null = null;

function finish(result: boolean) {
  visible.value = false;
  dialogState.value = null;
  const fn = pendingResolve;
  pendingResolve = null;
  fn?.(result);
}

/** Opens the global alert dialog and resolves true if the user confirms. */
export function confirmAlert(options: AlertDialogOptions): Promise<boolean> {
  if (pendingResolve) {
    finish(false);
  }
  // Avoid vue-i18n's heavy `t` generic in a non-component module (breaks tsc depth limit).
  const $t = (i18n.global as unknown as { t: (key: string) => string }).t;
  dialogState.value = {
    title: options.title,
    message: options.message ?? "",
    confirmLabel: options.confirmLabel ?? $t("common.confirm"),
    cancelLabel: options.cancelLabel ?? $t("common.cancel"),
    variant: options.variant ?? "danger",
  };
  visible.value = true;
  return new Promise<boolean>((resolve) => {
    pendingResolve = resolve;
  });
}

/** Used only by FxAlertDialogHost to wire confirm / cancel. */
export function useAlertDialogHost() {
  return {
    visible,
    dialogState,
    onConfirm: () => finish(true),
    onCancel: () => finish(false),
  };
}

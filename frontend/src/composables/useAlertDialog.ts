import { ref, shallowRef } from "vue";

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
    pendingResolve(false);
    pendingResolve = null;
    visible.value = false;
    dialogState.value = null;
  }
  dialogState.value = {
    title: options.title,
    message: options.message ?? "",
    confirmLabel: options.confirmLabel ?? "Confirm",
    cancelLabel: options.cancelLabel ?? "Cancel",
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

import { shallowRef, type ShallowRef } from "vue";

/** Handlers for item detail page (view + edit); CommandPalette invokes when available. */
export type ItemDetailCommandHandlers = {
  save?: () => void | Promise<void>;
  cancel?: () => void | Promise<void>;
  deleteItem?: () => void | Promise<void>;
  downloadQrPng?: () => void | Promise<void>;
  copyPageLink?: () => void | Promise<void>;
};

const handlers: ShallowRef<ItemDetailCommandHandlers | null> = shallowRef(null);

export function useItemDetailCommandHandlers(): ShallowRef<ItemDetailCommandHandlers | null> {
  return handlers;
}

export function setItemDetailCommandHandlers(h: ItemDetailCommandHandlers | null) {
  handlers.value = h;
}

import { shallowRef, type ShallowRef } from "vue";

export type ItemEditCommandHandlers = {
  save: () => void | Promise<void>;
  cancel: () => void | Promise<void>;
};

const handlers: ShallowRef<ItemEditCommandHandlers | null> = shallowRef(null);

/** Exposed for CommandPalette; same ref for all consumers. */
export function useItemEditCommandHandlers(): ShallowRef<ItemEditCommandHandlers | null> {
  return handlers;
}

export function setItemEditCommandHandlers(h: ItemEditCommandHandlers | null) {
  handlers.value = h;
}

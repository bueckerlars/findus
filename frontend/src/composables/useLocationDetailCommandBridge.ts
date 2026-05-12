import { shallowRef, type ShallowRef } from "vue";

export type LocationDetailCommandHandlers = {
  deleteLocation?: () => void | Promise<void>;
  downloadQrPng?: () => void | Promise<void>;
  copyPageLink?: () => void | Promise<void>;
};

const handlers: ShallowRef<LocationDetailCommandHandlers | null> = shallowRef(null);

export function useLocationDetailCommandHandlers(): ShallowRef<LocationDetailCommandHandlers | null> {
  return handlers;
}

export function setLocationDetailCommandHandlers(h: LocationDetailCommandHandlers | null) {
  handlers.value = h;
}

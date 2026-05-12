import { reactive } from "vue";

export type PendingCreate =
  | { kind: "item"; locationId?: string }
  | { kind: "location"; parentId?: string }
  | { kind: "label" };

const state = reactive({
  itemOpen: false,
  itemInitialLocationId: undefined as string | undefined,
  locationOpen: false,
  locationInitialParentId: undefined as string | undefined,
  labelOpen: false,
  pending: null as PendingCreate | null,
});

export function useCreateModals() {
  function closeAll() {
    state.itemOpen = false;
    state.locationOpen = false;
    state.labelOpen = false;
  }

  function openCreateItem(opts?: { locationId?: string }) {
    closeAll();
    state.itemInitialLocationId = opts?.locationId;
    state.itemOpen = true;
  }

  function openCreateLocation(opts?: { parentId?: string }) {
    closeAll();
    state.locationInitialParentId = opts?.parentId;
    state.locationOpen = true;
  }

  function openCreateLabel() {
    closeAll();
    state.labelOpen = true;
  }

  function setPending(p: PendingCreate) {
    state.pending = p;
  }

  function consumePending(): PendingCreate | null {
    const p = state.pending;
    state.pending = null;
    return p;
  }

  return {
    createModalState: state,
    openCreateItem,
    openCreateLocation,
    openCreateLabel,
    setPending,
    consumePending,
  };
}

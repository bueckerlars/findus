const LEGACY_KEY = "findus_items_view";

function storageKeyPart(storageKey: string) {
  return storageKey && storageKey.length ? storageKey : "default";
}

function itemsViewStorageKey(storageKey: string) {
  return "findus_items_view::" + storageKeyPart(storageKey);
}

export function readItemsViewMode(storageKey: string): "list" | "gallery" {
  const sk = itemsViewStorageKey(storageKey);
  try {
    const v = localStorage.getItem(sk);
    if (v === "gallery" || v === "list") return v;
    if (sk === "findus_items_view::page_items") {
      const leg = localStorage.getItem(LEGACY_KEY);
      if (leg === "gallery" || leg === "list") return leg;
    }
  } catch {
    /* ignore */
  }
  return "list";
}

export function writeItemsViewMode(storageKey: string, mode: "list" | "gallery") {
  try {
    localStorage.setItem(itemsViewStorageKey(storageKey), mode);
  } catch {
    /* ignore */
  }
}

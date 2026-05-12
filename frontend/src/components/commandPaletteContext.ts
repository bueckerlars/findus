import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { FxIconName } from "./FxSvg.vue";
import type { ItemDetailCommandHandlers } from "../composables/useItemDetailCommandBridge";
import type { LocationDetailCommandHandlers } from "../composables/useLocationDetailCommandBridge";

export type ContextCommand = {
  id: string;
  label: string;
  keywords: string;
  icon: FxIconName;
  href?: string;
  action?: string;
  /** Copy plain text after closing the palette (e.g. page URL). */
  clipboard?: string;
  /** Full navigation (non-SPA) after close — e.g. binary download. */
  externalHref?: string;
  /** Trigger file download for same-origin PNG URL. */
  downloadUrl?: string;
  downloadFilename?: string;
  /** Scoped list/gallery layout (see `itemsViewMode` composable keys). */
  itemsViewKey?: string;
  itemsViewMode?: "list" | "gallery";
  /** Open create modal with preset (same as page buttons). */
  createPreset?: { kind: "item"; locationId: string };
};

export type ContextPaletteDeps = {
  isAdmin: boolean;
  itemDetail: ItemDetailCommandHandlers | null;
  locationDetail: LocationDetailCommandHandlers | null;
};

export function buildContextCommands(route: RouteLocationNormalizedLoaded, d: ContextPaletteDeps): ContextCommand[] {
  const path = route.path;
  const { isAdmin: admin, itemDetail: ih, locationDetail: lh } = d;
  const out: ContextCommand[] = [];

  const itemsDetail = /^\/items\/([^/]+)$/.exec(path);
  if (itemsDetail?.[1] && itemsDetail[1] !== "new") {
    const id = itemsDetail[1];
    if (ih?.save && ih?.cancel) {
      out.push(
        {
          id: "ctx-item-save",
          label: "Save item",
          keywords: "save commit apply submit store persist",
          icon: "check",
          action: "item-detail:save",
        },
        {
          id: "ctx-item-cancel-edit",
          label: "Cancel editing",
          keywords: "cancel discard exit close abort revert discard changes",
          icon: "eye",
          action: "item-detail:cancel",
        },
      );
    } else if (admin) {
      out.push({
        id: "ctx-item-edit",
        label: "Edit item",
        keywords: "edit modify pencil form",
        icon: "pencilSquare",
        href: `/items/${id}?edit=1`,
      });
    }
    if (ih?.downloadQrPng) {
      out.push({
        id: "ctx-item-qr-dl",
        label: "Download item QR (PNG)",
        keywords: "qr code download png image print scan phone",
        icon: "qr",
        action: "item-detail:download-qr",
      });
    }
    if (ih?.copyPageLink) {
      out.push({
        id: "ctx-item-copy-link",
        label: "Copy link to this item",
        keywords: "copy url clipboard share link address",
        icon: "cube",
        action: "item-detail:copy-link",
      });
    }
    if (ih?.deleteItem) {
      out.push({
        id: "ctx-item-delete",
        label: "Delete item…",
        keywords: "delete remove trash destroy",
        icon: "trash",
        action: "item-detail:delete",
      });
    }
    out.push(
      {
        id: "ctx-item-photo-tab",
        label: "Open item photo (new tab)",
        keywords: "photo image picture open tab browser",
        icon: "photo",
        action: `tab:/items/${id}/photo`,
      },
      {
        id: "ctx-item-all",
        label: "All items",
        keywords: "items list inventory back",
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-item-search",
        label: "Search",
        keywords: "search find lookup filter",
        icon: "magnifyingGlass",
        href: "/search",
      },
    );
    return out;
  }

  const locDetail = /^\/locations\/([^/]+)$/.exec(path);
  if (locDetail?.[1] && locDetail[1] !== "new") {
    const id = locDetail[1];
    if (admin) {
      out.push(
        {
          id: "ctx-loc-edit",
          label: "Edit location",
          keywords: "edit modify rename",
          icon: "pencilSquare",
          href: `/locations/${id}/edit`,
        },
        {
          id: "ctx-loc-subloc",
          label: "Add sub-location",
          keywords: "new child sub place room",
          icon: "plus",
          href: `/locations/new?parent_id=${encodeURIComponent(id)}`,
        },
        {
          id: "ctx-loc-add-item",
          label: "Add item in this location",
          keywords: "new item create inventory here",
          icon: "cube",
          createPreset: { kind: "item", locationId: id },
        },
      );
    }
    if (lh?.downloadQrPng) {
      out.push({
        id: "ctx-loc-qr-dl",
        label: "Download location QR (PNG)",
        keywords: "qr code download png print scan phone",
        icon: "qr",
        action: "loc-detail:download-qr",
      });
    }
    if (lh?.copyPageLink) {
      out.push({
        id: "ctx-loc-copy-link",
        label: "Copy link to this location",
        keywords: "copy url clipboard share",
        icon: "cube",
        action: "loc-detail:copy-link",
      });
    }
    if (lh?.deleteLocation) {
      out.push({
        id: "ctx-loc-delete",
        label: "Delete location…",
        keywords: "delete remove trash empty",
        icon: "trash",
        action: "loc-detail:delete",
      });
    }
    out.push({
      id: "ctx-loc-all",
      label: "All locations",
      keywords: "locations places tree map back",
      icon: "mapPin",
      href: "/locations",
    });
    return out;
  }

  const locEdit = /^\/locations\/([^/]+)\/edit$/.exec(path);
  if (locEdit?.[1]) {
    const id = locEdit[1];
    out.push({
      id: "ctx-loc-view",
      label: "View location",
      keywords: "back cancel detail read",
      icon: "eye",
      href: `/locations/${id}`,
    });
    return out;
  }

  if (path === "/locations/new") {
    const pid = route.query.parent_id;
    if (typeof pid === "string" && pid.trim()) {
      out.push({
        id: "ctx-loc-parent",
        label: "Open parent location",
        keywords: "parent back navigate",
        icon: "mapPin",
        href: `/locations/${pid.trim()}`,
      });
    }
    out.push({
      id: "ctx-loc-new-all",
      label: "All locations",
      keywords: "locations list",
      icon: "mapPin",
      href: "/locations",
    });
    return out;
  }

  if (path === "/locations") {
    out.push(
      { id: "ctx-loc-list-search", label: "Search items", keywords: "search find", icon: "magnifyingGlass", href: "/search" },
      { id: "ctx-loc-list-items", label: "All items", keywords: "items inventory", icon: "cube", href: "/items" },
      { id: "ctx-loc-list-home", label: "Home", keywords: "dashboard", icon: "home", href: "/" },
    );
    return out;
  }

  if (path === "/items") {
    out.push(
      {
        id: "ctx-items-gallery",
        label: "Items: gallery view",
        keywords: "gallery grid cards layout",
        icon: "viewGrid",
        itemsViewKey: "page_items",
        itemsViewMode: "gallery",
        action: "items-view-mode",
      },
      {
        id: "ctx-items-list",
        label: "Items: list view",
        keywords: "list rows layout",
        icon: "viewList",
        itemsViewKey: "page_items",
        itemsViewMode: "list",
        action: "items-view-mode",
      },
      { id: "ctx-items-search", label: "Open search", keywords: "search find lookup filter", icon: "magnifyingGlass", href: "/search" },
      { id: "ctx-items-locations", label: "Locations", keywords: "places rooms map", icon: "mapPin", href: "/locations" },
    );
    return out;
  }

  if (path === "/") {
    out.push(
      {
        id: "ctx-home-recent-gallery",
        label: "Recent items: gallery",
        keywords: "home recent gallery grid",
        icon: "viewGrid",
        itemsViewKey: "home_recent_items",
        itemsViewMode: "gallery",
        action: "items-view-mode",
      },
      {
        id: "ctx-home-recent-list",
        label: "Recent items: list",
        keywords: "home recent list rows",
        icon: "viewList",
        itemsViewKey: "home_recent_items",
        itemsViewMode: "list",
        action: "items-view-mode",
      },
      { id: "ctx-home-search", label: "Open search", keywords: "search find lookup filter", icon: "magnifyingGlass", href: "/search" },
      { id: "ctx-home-items", label: "All items", keywords: "items inventory", icon: "cube", href: "/items" },
      { id: "ctx-home-locations", label: "All locations", keywords: "locations map places", icon: "mapPin", href: "/locations" },
      { id: "ctx-home-labels", label: "Labels", keywords: "tags categories", icon: "tag", href: "/labels" },
    );
    return out;
  }

  if (path === "/search") {
    out.push(
      {
        id: "ctx-search-focus",
        label: "Focus search field",
        keywords: "focus type query input cursor",
        icon: "magnifyingGlass",
        action: "focus:#q",
      },
      {
        id: "ctx-search-gallery",
        label: "Results: gallery view",
        keywords: "gallery grid layout",
        icon: "viewGrid",
        itemsViewKey: "page_search",
        itemsViewMode: "gallery",
        action: "items-view-mode",
      },
      {
        id: "ctx-search-list",
        label: "Results: list view",
        keywords: "list rows layout",
        icon: "viewList",
        itemsViewKey: "page_search",
        itemsViewMode: "list",
        action: "items-view-mode",
      },
      { id: "ctx-search-home", label: "Home", keywords: "dashboard start", icon: "home", href: "/" },
      { id: "ctx-search-items", label: "All items", keywords: "items browse", icon: "cube", href: "/items" },
    );
    return out;
  }

  if (path === "/labels") {
    if (admin) {
      out.push({
        id: "ctx-labels-new",
        label: "New label",
        keywords: "add create tag",
        icon: "plus",
        href: "/labels/new",
      });
    }
    out.push(
      { id: "ctx-labels-search", label: "Search items", keywords: "search find items", icon: "magnifyingGlass", href: "/search" },
      { id: "ctx-labels-home", label: "Home", icon: "home", keywords: "dashboard", href: "/" },
    );
    return out;
  }

  const labelEdit = /^\/labels\/([^/]+)\/edit$/.exec(path);
  if (labelEdit?.[1]) {
    const lid = labelEdit[1];
    out.push(
      { id: "ctx-label-all", label: "All labels", keywords: "labels list tags back", icon: "tag", href: "/labels" },
      { id: "ctx-label-search", label: "Search items", keywords: "search", icon: "magnifyingGlass", href: "/search" },
    );
    if (admin) {
      out.push({ id: "ctx-label-new2", label: "New label", keywords: "add create", icon: "plus", href: "/labels/new" });
    }
    return out;
  }

  if (path === "/labels/new") {
    out.push(
      { id: "ctx-label-new-all", label: "All labels", keywords: "labels list cancel back", icon: "tag", href: "/labels" },
      { id: "ctx-label-new-search", label: "Search items", icon: "magnifyingGlass", keywords: "search", href: "/search" },
    );
    return out;
  }

  if (path === "/admin/users") {
    out.push(
      {
        id: "ctx-admin-backup",
        label: "Download backup (ZIP)",
        keywords: "backup export zip archive download data",
        icon: "gear",
        externalHref: "/admin/backup.zip",
      },
      {
        id: "ctx-admin-tpl",
        label: "Templates",
        keywords: "item template fields editor",
        icon: "gear",
        href: "/admin/templates",
      },
      { id: "ctx-admin-home", label: "Home", keywords: "dashboard", icon: "home", href: "/" },
    );
    return out;
  }

  if (path === "/admin/templates") {
    out.push(
      {
        id: "ctx-tpl-users",
        label: "Users",
        keywords: "admin invites user management",
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-tpl-new",
        label: "New template",
        keywords: "add create",
        icon: "plus",
        href: "/admin/templates/new",
      },
      { id: "ctx-tpl-home", label: "Home", keywords: "dashboard", icon: "home", href: "/" },
    );
    return out;
  }

  if (path === "/admin/templates/new") {
    out.push({
      id: "ctx-tpl-new-list",
      label: "All templates",
      keywords: "templates list back cancel",
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  const tplEdit = /^\/admin\/templates\/([^/]+)\/edit$/.exec(path);
  if (tplEdit?.[1]) {
    out.push({
      id: "ctx-tpl-edit-list",
      label: "All templates",
      keywords: "templates list back cancel",
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  if (path === "/items/new") {
    out.push({
      id: "ctx-itemform-items",
      label: "All items",
      keywords: "items list cancel back",
      icon: "cube",
      href: "/items",
    });
    return out;
  }

  if (path === "/profile") {
    out.push(
      { id: "ctx-profile-home", label: "Home", keywords: "dashboard", icon: "home", href: "/" },
      { id: "ctx-profile-search", label: "Search items", keywords: "search", icon: "magnifyingGlass", href: "/search" },
      {
        id: "ctx-profile-user",
        label: "Focus username",
        keywords: "account name user edit",
        icon: "users",
        action: "focus:#pu",
      },
      {
        id: "ctx-profile-email",
        label: "Focus email",
        keywords: "mail address",
        icon: "users",
        action: "focus:#pe",
      },
      {
        id: "ctx-profile-avatar",
        label: "Choose avatar image",
        keywords: "photo picture upload image",
        icon: "photo",
        action: "focus:#pav",
      },
    );
    return out;
  }

  return out;
}

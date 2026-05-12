import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { ComposerTranslation } from "vue-i18n";
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
  /** Open create modal with preset (same as page buttons). */
  createPreset?: { kind: "item"; locationId: string };
};

export type ContextPaletteDeps = {
  isAdmin: boolean;
  itemDetail: ItemDetailCommandHandlers | null;
  locationDetail: LocationDetailCommandHandlers | null;
};

function ctxCmd(t: ComposerTranslation, id: string, part: "l" | "k"): string {
  const suffix = id.replace(/-/g, "_") + "_" + part;
  return t("cp." + suffix);
}

export function buildContextCommands(
  route: RouteLocationNormalizedLoaded,
  d: ContextPaletteDeps,
  t: ComposerTranslation,
): ContextCommand[] {
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
          label: ctxCmd(t, "ctx-item-save", "l"),
          keywords: ctxCmd(t, "ctx-item-save", "k"),
          icon: "check",
          action: "item-detail:save",
        },
        {
          id: "ctx-item-cancel-edit",
          label: ctxCmd(t, "ctx-item-cancel-edit", "l"),
          keywords: ctxCmd(t, "ctx-item-cancel-edit", "k"),
          icon: "eye",
          action: "item-detail:cancel",
        },
      );
    } else if (admin) {
      out.push({
        id: "ctx-item-edit",
        label: ctxCmd(t, "ctx-item-edit", "l"),
        keywords: ctxCmd(t, "ctx-item-edit", "k"),
        icon: "pencilSquare",
        href: `/items/${id}?edit=1`,
      });
    }
    if (ih?.downloadQrPng) {
      out.push({
        id: "ctx-item-qr-dl",
        label: ctxCmd(t, "ctx-item-qr-dl", "l"),
        keywords: ctxCmd(t, "ctx-item-qr-dl", "k"),
        icon: "qr",
        action: "item-detail:download-qr",
      });
    }
    if (ih?.copyPageLink) {
      out.push({
        id: "ctx-item-copy-link",
        label: ctxCmd(t, "ctx-item-copy-link", "l"),
        keywords: ctxCmd(t, "ctx-item-copy-link", "k"),
        icon: "cube",
        action: "item-detail:copy-link",
      });
    }
    if (ih?.deleteItem) {
      out.push({
        id: "ctx-item-delete",
        label: ctxCmd(t, "ctx-item-delete", "l"),
        keywords: ctxCmd(t, "ctx-item-delete", "k"),
        icon: "trash",
        action: "item-detail:delete",
      });
    }
    out.push(
      {
        id: "ctx-item-photo-tab",
        label: ctxCmd(t, "ctx-item-photo-tab", "l"),
        keywords: ctxCmd(t, "ctx-item-photo-tab", "k"),
        icon: "photo",
        action: `tab:/items/${id}/photo`,
      },
      {
        id: "ctx-item-all",
        label: ctxCmd(t, "ctx-item-all", "l"),
        keywords: ctxCmd(t, "ctx-item-all", "k"),
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-item-search",
        label: ctxCmd(t, "ctx-item-search", "l"),
        keywords: ctxCmd(t, "ctx-item-search", "k"),
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
          label: ctxCmd(t, "ctx-loc-edit", "l"),
          keywords: ctxCmd(t, "ctx-loc-edit", "k"),
          icon: "pencilSquare",
          href: `/locations/${id}/edit`,
        },
        {
          id: "ctx-loc-subloc",
          label: ctxCmd(t, "ctx-loc-subloc", "l"),
          keywords: ctxCmd(t, "ctx-loc-subloc", "k"),
          icon: "plus",
          href: `/locations/new?parent_id=${encodeURIComponent(id)}`,
        },
        {
          id: "ctx-loc-add-item",
          label: ctxCmd(t, "ctx-loc-add-item", "l"),
          keywords: ctxCmd(t, "ctx-loc-add-item", "k"),
          icon: "cube",
          createPreset: { kind: "item", locationId: id },
        },
      );
    }
    if (lh?.downloadQrPng) {
      out.push({
        id: "ctx-loc-qr-dl",
        label: ctxCmd(t, "ctx-loc-qr-dl", "l"),
        keywords: ctxCmd(t, "ctx-loc-qr-dl", "k"),
        icon: "qr",
        action: "loc-detail:download-qr",
      });
    }
    if (lh?.copyPageLink) {
      out.push({
        id: "ctx-loc-copy-link",
        label: ctxCmd(t, "ctx-loc-copy-link", "l"),
        keywords: ctxCmd(t, "ctx-loc-copy-link", "k"),
        icon: "cube",
        action: "loc-detail:copy-link",
      });
    }
    if (lh?.deleteLocation) {
      out.push({
        id: "ctx-loc-delete",
        label: ctxCmd(t, "ctx-loc-delete", "l"),
        keywords: ctxCmd(t, "ctx-loc-delete", "k"),
        icon: "trash",
        action: "loc-detail:delete",
      });
    }
    out.push({
      id: "ctx-loc-all",
      label: ctxCmd(t, "ctx-loc-all", "l"),
      keywords: ctxCmd(t, "ctx-loc-all", "k"),
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
      label: ctxCmd(t, "ctx-loc-view", "l"),
      keywords: ctxCmd(t, "ctx-loc-view", "k"),
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
        label: ctxCmd(t, "ctx-loc-parent", "l"),
        keywords: ctxCmd(t, "ctx-loc-parent", "k"),
        icon: "mapPin",
        href: `/locations/${pid.trim()}`,
      });
    }
    out.push({
      id: "ctx-loc-new-all",
      label: ctxCmd(t, "ctx-loc-new-all", "l"),
      keywords: ctxCmd(t, "ctx-loc-new-all", "k"),
      icon: "mapPin",
      href: "/locations",
    });
    return out;
  }

  if (path === "/locations") {
    out.push(
      {
        id: "ctx-loc-list-search",
        label: ctxCmd(t, "ctx-loc-list-search", "l"),
        keywords: ctxCmd(t, "ctx-loc-list-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-loc-list-items",
        label: ctxCmd(t, "ctx-loc-list-items", "l"),
        keywords: ctxCmd(t, "ctx-loc-list-items", "k"),
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-loc-list-home",
        label: ctxCmd(t, "ctx-loc-list-home", "l"),
        keywords: ctxCmd(t, "ctx-loc-list-home", "k"),
        icon: "home",
        href: "/",
      },
    );
    return out;
  }

  if (path === "/items") {
    out.push(
      {
        id: "ctx-items-search",
        label: ctxCmd(t, "ctx-items-search", "l"),
        keywords: ctxCmd(t, "ctx-items-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-items-locations",
        label: ctxCmd(t, "ctx-items-locations", "l"),
        keywords: ctxCmd(t, "ctx-items-locations", "k"),
        icon: "mapPin",
        href: "/locations",
      },
    );
    return out;
  }

  if (path === "/") {
    out.push(
      {
        id: "ctx-home-search",
        label: ctxCmd(t, "ctx-home-search", "l"),
        keywords: ctxCmd(t, "ctx-home-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-home-items",
        label: ctxCmd(t, "ctx-home-items", "l"),
        keywords: ctxCmd(t, "ctx-home-items", "k"),
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-home-locations",
        label: ctxCmd(t, "ctx-home-locations", "l"),
        keywords: ctxCmd(t, "ctx-home-locations", "k"),
        icon: "mapPin",
        href: "/locations",
      },
      {
        id: "ctx-home-labels",
        label: ctxCmd(t, "ctx-home-labels", "l"),
        keywords: ctxCmd(t, "ctx-home-labels", "k"),
        icon: "tag",
        href: "/labels",
      },
    );
    return out;
  }

  if (path === "/search") {
    out.push(
      {
        id: "ctx-search-focus",
        label: ctxCmd(t, "ctx-search-focus", "l"),
        keywords: ctxCmd(t, "ctx-search-focus", "k"),
        icon: "magnifyingGlass",
        action: "focus:#q",
      },
      {
        id: "ctx-search-home",
        label: ctxCmd(t, "ctx-search-home", "l"),
        keywords: ctxCmd(t, "ctx-search-home", "k"),
        icon: "home",
        href: "/",
      },
      {
        id: "ctx-search-items",
        label: ctxCmd(t, "ctx-search-items", "l"),
        keywords: ctxCmd(t, "ctx-search-items", "k"),
        icon: "cube",
        href: "/items",
      },
    );
    return out;
  }

  if (path === "/labels") {
    if (admin) {
      out.push({
        id: "ctx-labels-new",
        label: ctxCmd(t, "ctx-labels-new", "l"),
        keywords: ctxCmd(t, "ctx-labels-new", "k"),
        icon: "plus",
        href: "/labels/new",
      });
    }
    out.push(
      {
        id: "ctx-labels-search",
        label: ctxCmd(t, "ctx-labels-search", "l"),
        keywords: ctxCmd(t, "ctx-labels-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-labels-home",
        label: ctxCmd(t, "ctx-labels-home", "l"),
        keywords: ctxCmd(t, "ctx-labels-home", "k"),
        icon: "home",
        href: "/",
      },
    );
    return out;
  }

  const labelEdit = /^\/labels\/([^/]+)\/edit$/.exec(path);
  if (labelEdit?.[1]) {
    out.push(
      {
        id: "ctx-label-all",
        label: ctxCmd(t, "ctx-label-all", "l"),
        keywords: ctxCmd(t, "ctx-label-all", "k"),
        icon: "tag",
        href: "/labels",
      },
      {
        id: "ctx-label-search",
        label: ctxCmd(t, "ctx-label-search", "l"),
        keywords: ctxCmd(t, "ctx-label-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
    );
    if (admin) {
      out.push({
        id: "ctx-label-new2",
        label: ctxCmd(t, "ctx-label-new2", "l"),
        keywords: ctxCmd(t, "ctx-label-new2", "k"),
        icon: "plus",
        href: "/labels/new",
      });
    }
    return out;
  }

  if (path === "/labels/new") {
    out.push(
      {
        id: "ctx-label-new-all",
        label: ctxCmd(t, "ctx-label-new-all", "l"),
        keywords: ctxCmd(t, "ctx-label-new-all", "k"),
        icon: "tag",
        href: "/labels",
      },
      {
        id: "ctx-label-new-search",
        label: ctxCmd(t, "ctx-label-new-search", "l"),
        keywords: ctxCmd(t, "ctx-label-new-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
    );
    return out;
  }

  if (path === "/admin/users") {
    out.push(
      {
        id: "ctx-admin-settings",
        label: ctxCmd(t, "ctx-admin-settings", "l"),
        keywords: ctxCmd(t, "ctx-admin-settings", "k"),
        icon: "gear",
        href: "/admin/settings",
      },
      {
        id: "ctx-admin-backup",
        label: ctxCmd(t, "ctx-admin-backup", "l"),
        keywords: ctxCmd(t, "ctx-admin-backup", "k"),
        icon: "gear",
        externalHref: "/admin/backup.zip",
      },
      {
        id: "ctx-admin-tpl",
        label: ctxCmd(t, "ctx-admin-tpl", "l"),
        keywords: ctxCmd(t, "ctx-admin-tpl", "k"),
        icon: "gear",
        href: "/admin/templates",
      },
      {
        id: "ctx-admin-home",
        label: ctxCmd(t, "ctx-admin-home", "l"),
        keywords: ctxCmd(t, "ctx-admin-home", "k"),
        icon: "home",
        href: "/",
      },
    );
    return out;
  }

  if (path === "/admin/settings") {
    out.push(
      {
        id: "ctx-settings-users",
        label: ctxCmd(t, "ctx-settings-users", "l"),
        keywords: ctxCmd(t, "ctx-settings-users", "k"),
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-settings-backup",
        label: ctxCmd(t, "ctx-settings-backup", "l"),
        keywords: ctxCmd(t, "ctx-settings-backup", "k"),
        icon: "gear",
        externalHref: "/admin/backup.zip",
      },
      {
        id: "ctx-settings-tpl",
        label: ctxCmd(t, "ctx-settings-tpl", "l"),
        keywords: ctxCmd(t, "ctx-settings-tpl", "k"),
        icon: "gear",
        href: "/admin/templates",
      },
      {
        id: "ctx-settings-home",
        label: ctxCmd(t, "ctx-settings-home", "l"),
        keywords: ctxCmd(t, "ctx-settings-home", "k"),
        icon: "home",
        href: "/",
      },
    );
    return out;
  }

  if (path === "/admin/templates") {
    out.push(
      {
        id: "ctx-tpl-settings",
        label: ctxCmd(t, "ctx-tpl-settings", "l"),
        keywords: ctxCmd(t, "ctx-tpl-settings", "k"),
        icon: "gear",
        href: "/admin/settings",
      },
      {
        id: "ctx-tpl-users",
        label: ctxCmd(t, "ctx-tpl-users", "l"),
        keywords: ctxCmd(t, "ctx-tpl-users", "k"),
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-tpl-new",
        label: ctxCmd(t, "ctx-tpl-new", "l"),
        keywords: ctxCmd(t, "ctx-tpl-new", "k"),
        icon: "plus",
        href: "/admin/templates/new",
      },
      {
        id: "ctx-tpl-home",
        label: ctxCmd(t, "ctx-tpl-home", "l"),
        keywords: ctxCmd(t, "ctx-tpl-home", "k"),
        icon: "home",
        href: "/",
      },
    );
    return out;
  }

  if (path === "/admin/templates/new") {
    out.push({
      id: "ctx-tpl-new-list",
      label: ctxCmd(t, "ctx-tpl-new-list", "l"),
      keywords: ctxCmd(t, "ctx-tpl-new-list", "k"),
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  const tplEdit = /^\/admin\/templates\/([^/]+)\/edit$/.exec(path);
  if (tplEdit?.[1]) {
    out.push({
      id: "ctx-tpl-edit-list",
      label: ctxCmd(t, "ctx-tpl-edit-list", "l"),
      keywords: ctxCmd(t, "ctx-tpl-edit-list", "k"),
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  if (path === "/items/new") {
    out.push({
      id: "ctx-itemform-items",
      label: ctxCmd(t, "ctx-itemform-items", "l"),
      keywords: ctxCmd(t, "ctx-itemform-items", "k"),
      icon: "cube",
      href: "/items",
    });
    return out;
  }

  if (path === "/profile") {
    out.push(
      {
        id: "ctx-profile-home",
        label: ctxCmd(t, "ctx-profile-home", "l"),
        keywords: ctxCmd(t, "ctx-profile-home", "k"),
        icon: "home",
        href: "/",
      },
      {
        id: "ctx-profile-search",
        label: ctxCmd(t, "ctx-profile-search", "l"),
        keywords: ctxCmd(t, "ctx-profile-search", "k"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-profile-user",
        label: ctxCmd(t, "ctx-profile-user", "l"),
        keywords: ctxCmd(t, "ctx-profile-user", "k"),
        icon: "users",
        action: "focus:#pu",
      },
      {
        id: "ctx-profile-email",
        label: ctxCmd(t, "ctx-profile-email", "l"),
        keywords: ctxCmd(t, "ctx-profile-email", "k"),
        icon: "users",
        action: "focus:#pe",
      },
      {
        id: "ctx-profile-avatar",
        label: ctxCmd(t, "ctx-profile-avatar", "l"),
        keywords: ctxCmd(t, "ctx-profile-avatar", "k"),
        icon: "photo",
        action: "focus:#pav",
      },
    );
    return out;
  }

  return out;
}

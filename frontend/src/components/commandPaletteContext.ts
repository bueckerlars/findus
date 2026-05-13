import type { RouteLocationNormalizedLoaded } from "vue-router";
import type { ComposerTranslation } from "vue-i18n";
import { allLocalesSearchBlob } from "../utils/commandPaletteAllLocaleKeywords";
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
  canEditItems: boolean;
  canEditLocations: boolean;
  canEditLabels: boolean;
  itemDetail: ItemDetailCommandHandlers | null;
  locationDetail: LocationDetailCommandHandlers | null;
};

function ctxCmd(t: ComposerTranslation, id: string, part: "l" | "k"): string {
  const suffix = id.replace(/-/g, "_") + "_" + part;
  return t("cp." + suffix);
}

function ctxCmdKeywords(t: ComposerTranslation, id: string): string {
  const suffix = id.replace(/-/g, "_");
  const keys = [`cp.${suffix}_k`, `cp.${suffix}_l`] as const;
  return [ctxCmd(t, id, "k"), allLocalesSearchBlob(keys)].filter(Boolean).join(" ");
}

export function buildContextCommands(
  route: RouteLocationNormalizedLoaded,
  d: ContextPaletteDeps,
  t: ComposerTranslation,
): ContextCommand[] {
  const path = route.path;
  const { canEditItems, canEditLocations, canEditLabels, itemDetail: ih, locationDetail: lh } = d;
  const out: ContextCommand[] = [];

  const itemsDetail = /^\/items\/([^/]+)$/.exec(path);
  if (itemsDetail?.[1] && itemsDetail[1] !== "new") {
    const id = itemsDetail[1];
    if (ih?.save && ih?.cancel) {
      out.push(
        {
          id: "ctx-item-save",
          label: ctxCmd(t, "ctx-item-save", "l"),
          keywords: ctxCmdKeywords(t, "ctx-item-save"),
          icon: "check",
          action: "item-detail:save",
        },
        {
          id: "ctx-item-cancel-edit",
          label: ctxCmd(t, "ctx-item-cancel-edit", "l"),
          keywords: ctxCmdKeywords(t, "ctx-item-cancel-edit"),
          icon: "eye",
          action: "item-detail:cancel",
        },
      );
    } else if (canEditItems) {
      out.push({
        id: "ctx-item-edit",
        label: ctxCmd(t, "ctx-item-edit", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-edit"),
        icon: "pencilSquare",
        href: `/items/${id}?edit=1`,
      });
    }
    if (ih?.downloadQrPng) {
      out.push({
        id: "ctx-item-qr-dl",
        label: ctxCmd(t, "ctx-item-qr-dl", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-qr-dl"),
        icon: "qr",
        action: "item-detail:download-qr",
      });
    }
    if (ih?.copyPageLink) {
      out.push({
        id: "ctx-item-copy-link",
        label: ctxCmd(t, "ctx-item-copy-link", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-copy-link"),
        icon: "cube",
        action: "item-detail:copy-link",
      });
    }
    if (ih?.deleteItem) {
      out.push({
        id: "ctx-item-delete",
        label: ctxCmd(t, "ctx-item-delete", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-delete"),
        icon: "trash",
        action: "item-detail:delete",
      });
    }
    out.push(
      {
        id: "ctx-item-photo-tab",
        label: ctxCmd(t, "ctx-item-photo-tab", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-photo-tab"),
        icon: "photo",
        action: `tab:/items/${id}/photo`,
      },
      {
        id: "ctx-item-all",
        label: ctxCmd(t, "ctx-item-all", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-all"),
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-item-search",
        label: ctxCmd(t, "ctx-item-search", "l"),
        keywords: ctxCmdKeywords(t, "ctx-item-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
    );
    return out;
  }

  const locDetail = /^\/locations\/([^/]+)$/.exec(path);
  if (locDetail?.[1] && locDetail[1] !== "new") {
    const id = locDetail[1];
    if (canEditLocations) {
      out.push(
        {
          id: "ctx-loc-edit",
          label: ctxCmd(t, "ctx-loc-edit", "l"),
          keywords: ctxCmdKeywords(t, "ctx-loc-edit"),
          icon: "pencilSquare",
          href: `/locations/${id}/edit`,
        },
        {
          id: "ctx-loc-subloc",
          label: ctxCmd(t, "ctx-loc-subloc", "l"),
          keywords: ctxCmdKeywords(t, "ctx-loc-subloc"),
          icon: "plus",
          href: `/locations/new?parent_id=${encodeURIComponent(id)}`,
        },
      );
    }
    if (canEditItems) {
      out.push({
        id: "ctx-loc-add-item",
        label: ctxCmd(t, "ctx-loc-add-item", "l"),
        keywords: ctxCmdKeywords(t, "ctx-loc-add-item"),
        icon: "cube",
        createPreset: { kind: "item", locationId: id },
      });
    }
    if (lh?.downloadQrPng) {
      out.push({
        id: "ctx-loc-qr-dl",
        label: ctxCmd(t, "ctx-loc-qr-dl", "l"),
        keywords: ctxCmdKeywords(t, "ctx-loc-qr-dl"),
        icon: "qr",
        action: "loc-detail:download-qr",
      });
    }
    if (lh?.copyPageLink) {
      out.push({
        id: "ctx-loc-copy-link",
        label: ctxCmd(t, "ctx-loc-copy-link", "l"),
        keywords: ctxCmdKeywords(t, "ctx-loc-copy-link"),
        icon: "cube",
        action: "loc-detail:copy-link",
      });
    }
    if (lh?.deleteLocation) {
      out.push({
        id: "ctx-loc-delete",
        label: ctxCmd(t, "ctx-loc-delete", "l"),
        keywords: ctxCmdKeywords(t, "ctx-loc-delete"),
        icon: "trash",
        action: "loc-detail:delete",
      });
    }
    out.push({
      id: "ctx-loc-all",
      label: ctxCmd(t, "ctx-loc-all", "l"),
      keywords: ctxCmdKeywords(t, "ctx-loc-all"),
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
      keywords: ctxCmdKeywords(t, "ctx-loc-view"),
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
        keywords: ctxCmdKeywords(t, "ctx-loc-parent"),
        icon: "mapPin",
        href: `/locations/${pid.trim()}`,
      });
    }
    out.push({
      id: "ctx-loc-new-all",
      label: ctxCmd(t, "ctx-loc-new-all", "l"),
      keywords: ctxCmdKeywords(t, "ctx-loc-new-all"),
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
        keywords: ctxCmdKeywords(t, "ctx-loc-list-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-loc-list-items",
        label: ctxCmd(t, "ctx-loc-list-items", "l"),
        keywords: ctxCmdKeywords(t, "ctx-loc-list-items"),
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-loc-list-home",
        label: ctxCmd(t, "ctx-loc-list-home", "l"),
        keywords: ctxCmdKeywords(t, "ctx-loc-list-home"),
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
        keywords: ctxCmdKeywords(t, "ctx-items-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-items-locations",
        label: ctxCmd(t, "ctx-items-locations", "l"),
        keywords: ctxCmdKeywords(t, "ctx-items-locations"),
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
        keywords: ctxCmdKeywords(t, "ctx-home-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-home-items",
        label: ctxCmd(t, "ctx-home-items", "l"),
        keywords: ctxCmdKeywords(t, "ctx-home-items"),
        icon: "cube",
        href: "/items",
      },
      {
        id: "ctx-home-locations",
        label: ctxCmd(t, "ctx-home-locations", "l"),
        keywords: ctxCmdKeywords(t, "ctx-home-locations"),
        icon: "mapPin",
        href: "/locations",
      },
      {
        id: "ctx-home-labels",
        label: ctxCmd(t, "ctx-home-labels", "l"),
        keywords: ctxCmdKeywords(t, "ctx-home-labels"),
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
        keywords: ctxCmdKeywords(t, "ctx-search-focus"),
        icon: "magnifyingGlass",
        action: "focus:#q",
      },
      {
        id: "ctx-search-home",
        label: ctxCmd(t, "ctx-search-home", "l"),
        keywords: ctxCmdKeywords(t, "ctx-search-home"),
        icon: "home",
        href: "/",
      },
      {
        id: "ctx-search-items",
        label: ctxCmd(t, "ctx-search-items", "l"),
        keywords: ctxCmdKeywords(t, "ctx-search-items"),
        icon: "cube",
        href: "/items",
      },
    );
    return out;
  }

  if (path === "/labels") {
    if (canEditLabels) {
      out.push({
        id: "ctx-labels-new",
        label: ctxCmd(t, "ctx-labels-new", "l"),
        keywords: ctxCmdKeywords(t, "ctx-labels-new"),
        icon: "plus",
        href: "/labels/new",
      });
    }
    out.push(
      {
        id: "ctx-labels-search",
        label: ctxCmd(t, "ctx-labels-search", "l"),
        keywords: ctxCmdKeywords(t, "ctx-labels-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-labels-home",
        label: ctxCmd(t, "ctx-labels-home", "l"),
        keywords: ctxCmdKeywords(t, "ctx-labels-home"),
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
        keywords: ctxCmdKeywords(t, "ctx-label-all"),
        icon: "tag",
        href: "/labels",
      },
      {
        id: "ctx-label-search",
        label: ctxCmd(t, "ctx-label-search", "l"),
        keywords: ctxCmdKeywords(t, "ctx-label-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
    );
    if (canEditLabels) {
      out.push({
        id: "ctx-label-new2",
        label: ctxCmd(t, "ctx-label-new2", "l"),
        keywords: ctxCmdKeywords(t, "ctx-label-new2"),
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
        keywords: ctxCmdKeywords(t, "ctx-label-new-all"),
        icon: "tag",
        href: "/labels",
      },
      {
        id: "ctx-label-new-search",
        label: ctxCmd(t, "ctx-label-new-search", "l"),
        keywords: ctxCmdKeywords(t, "ctx-label-new-search"),
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
        keywords: ctxCmdKeywords(t, "ctx-admin-settings"),
        icon: "gear",
        href: "/admin/settings",
      },
      {
        id: "ctx-admin-backup",
        label: ctxCmd(t, "ctx-admin-backup", "l"),
        keywords: ctxCmdKeywords(t, "ctx-admin-backup"),
        icon: "gear",
        externalHref: "/admin/backup.zip",
      },
      {
        id: "ctx-admin-tpl",
        label: ctxCmd(t, "ctx-admin-tpl", "l"),
        keywords: ctxCmdKeywords(t, "ctx-admin-tpl"),
        icon: "gear",
        href: "/admin/templates",
      },
      {
        id: "ctx-admin-label-gen",
        label: ctxCmd(t, "ctx-admin-label-gen", "l"),
        keywords: ctxCmdKeywords(t, "ctx-admin-label-gen"),
        icon: "qr",
        href: "/tools",
      },
      {
        id: "ctx-admin-home",
        label: ctxCmd(t, "ctx-admin-home", "l"),
        keywords: ctxCmdKeywords(t, "ctx-admin-home"),
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
        keywords: ctxCmdKeywords(t, "ctx-settings-users"),
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-settings-backup",
        label: ctxCmd(t, "ctx-settings-backup", "l"),
        keywords: ctxCmdKeywords(t, "ctx-settings-backup"),
        icon: "gear",
        externalHref: "/admin/backup.zip",
      },
      {
        id: "ctx-settings-tpl",
        label: ctxCmd(t, "ctx-settings-tpl", "l"),
        keywords: ctxCmdKeywords(t, "ctx-settings-tpl"),
        icon: "gear",
        href: "/admin/templates",
      },
      {
        id: "ctx-settings-label-gen",
        label: ctxCmd(t, "ctx-settings-label-gen", "l"),
        keywords: ctxCmdKeywords(t, "ctx-settings-label-gen"),
        icon: "qr",
        href: "/tools",
      },
      {
        id: "ctx-settings-home",
        label: ctxCmd(t, "ctx-settings-home", "l"),
        keywords: ctxCmdKeywords(t, "ctx-settings-home"),
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
        keywords: ctxCmdKeywords(t, "ctx-tpl-settings"),
        icon: "gear",
        href: "/admin/settings",
      },
      {
        id: "ctx-tpl-users",
        label: ctxCmd(t, "ctx-tpl-users", "l"),
        keywords: ctxCmdKeywords(t, "ctx-tpl-users"),
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-tpl-new",
        label: ctxCmd(t, "ctx-tpl-new", "l"),
        keywords: ctxCmdKeywords(t, "ctx-tpl-new"),
        icon: "plus",
        href: "/admin/templates/new",
      },
      {
        id: "ctx-tpl-label-gen",
        label: ctxCmd(t, "ctx-tpl-label-gen", "l"),
        keywords: ctxCmdKeywords(t, "ctx-tpl-label-gen"),
        icon: "qr",
        href: "/tools",
      },
      {
        id: "ctx-tpl-home",
        label: ctxCmd(t, "ctx-tpl-home", "l"),
        keywords: ctxCmdKeywords(t, "ctx-tpl-home"),
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
      keywords: ctxCmdKeywords(t, "ctx-tpl-new-list"),
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  if (path === "/tools") {
    out.push(
      {
        id: "ctx-labelgen-settings",
        label: ctxCmd(t, "ctx-labelgen-settings", "l"),
        keywords: ctxCmdKeywords(t, "ctx-labelgen-settings"),
        icon: "gear",
        href: "/admin/settings",
      },
      {
        id: "ctx-labelgen-users",
        label: ctxCmd(t, "ctx-labelgen-users", "l"),
        keywords: ctxCmdKeywords(t, "ctx-labelgen-users"),
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-labelgen-tpl",
        label: ctxCmd(t, "ctx-labelgen-tpl", "l"),
        keywords: ctxCmdKeywords(t, "ctx-labelgen-tpl"),
        icon: "gear",
        href: "/admin/templates",
      },
    );
    return out;
  }

  const tplEdit = /^\/admin\/templates\/([^/]+)\/edit$/.exec(path);
  if (tplEdit?.[1]) {
    out.push({
      id: "ctx-tpl-edit-list",
      label: ctxCmd(t, "ctx-tpl-edit-list", "l"),
      keywords: ctxCmdKeywords(t, "ctx-tpl-edit-list"),
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  if (path === "/items/new") {
    out.push({
      id: "ctx-itemform-items",
      label: ctxCmd(t, "ctx-itemform-items", "l"),
      keywords: ctxCmdKeywords(t, "ctx-itemform-items"),
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
        keywords: ctxCmdKeywords(t, "ctx-profile-home"),
        icon: "home",
        href: "/",
      },
      {
        id: "ctx-profile-search",
        label: ctxCmd(t, "ctx-profile-search", "l"),
        keywords: ctxCmdKeywords(t, "ctx-profile-search"),
        icon: "magnifyingGlass",
        href: "/search",
      },
      {
        id: "ctx-profile-user",
        label: ctxCmd(t, "ctx-profile-user", "l"),
        keywords: ctxCmdKeywords(t, "ctx-profile-user"),
        icon: "users",
        action: "focus:#pu",
      },
      {
        id: "ctx-profile-email",
        label: ctxCmd(t, "ctx-profile-email", "l"),
        keywords: ctxCmdKeywords(t, "ctx-profile-email"),
        icon: "users",
        action: "focus:#pe",
      },
      {
        id: "ctx-profile-avatar",
        label: ctxCmd(t, "ctx-profile-avatar", "l"),
        keywords: ctxCmdKeywords(t, "ctx-profile-avatar"),
        icon: "photo",
        action: "focus:#pav",
      },
    );
    return out;
  }

  return out;
}

import { LOCALE_MESSAGES, SUPPORTED_LOCALES } from "../locale/bundle";

function readLeaf(messages: Record<string, unknown>, dotKey: string): string | undefined {
  const parts = dotKey.split(".");
  let cur: unknown = messages;
  for (const p of parts) {
    if (!cur || typeof cur !== "object" || Array.isArray(cur)) return undefined;
    const o = cur as Record<string, unknown>;
    if (!(p in o)) return undefined;
    cur = o[p];
  }
  return typeof cur === "string" ? cur : undefined;
}

/** Space-separated strings for `keys` from every loaded UI locale (for command-palette substring search). */
export function allLocalesSearchBlob(keys: readonly string[]): string {
  const parts: string[] = [];
  for (const code of SUPPORTED_LOCALES) {
    const messages = LOCALE_MESSAGES[code] as Record<string, unknown>;
    for (const key of keys) {
      const s = readLeaf(messages, key);
      if (s?.trim()) parts.push(s);
    }
  }
  return parts.join(" ");
}

/** Active sidebar / nav highlight (matches legacy `navActive`). */
export function navActive(currentPath: string, prefix: string): boolean {
  const p = currentPath || "/";
  if (prefix === "/") {
    return p === "/" || p === "";
  }
  return p === prefix || p.startsWith(prefix + "/");
}

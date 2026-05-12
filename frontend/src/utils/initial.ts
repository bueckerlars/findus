export function usernameInitial(username: string): string {
  const u = (username || "").trim();
  if (!u) return "?";
  return u.slice(0, 1).toUpperCase();
}

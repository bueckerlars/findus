export function csrfToken(): string {
  const m = document.cookie.match(/(?:^|; )findus_csrf=([^;]*)/);
  return m ? decodeURIComponent(m[1]) : "";
}

export async function postJson<T>(path: string, body: unknown): Promise<T> {
  return api<T>(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export async function patchJson<T>(path: string, body: unknown): Promise<T> {
  return api<T>(path, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const method = (init.method || "GET").toUpperCase();
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    const t = csrfToken();
    if (t) headers.set("X-CSRF-Token", t);
  }
  const res = await fetch(path, {
    ...init,
    credentials: "same-origin",
    headers,
  });
  if (res.status === 401) {
    const skipRedirect =
      path.startsWith("/api/me") ||
      path.startsWith("/api/auth/login") ||
      path.startsWith("/api/auth/register");
    if (!skipRedirect) {
      const next = encodeURIComponent(location.pathname + location.search);
      window.location.href = "/login?next=" + next;
    }
    throw new Error("unauthorized");
  }
  const ct = res.headers.get("Content-Type") || "";
  if (!res.ok) {
    let msg = res.statusText;
    if (ct.includes("application/json")) {
      try {
        const j = (await res.json()) as { error?: string };
        if (j.error) msg = j.error;
      } catch {
        /* ignore */
      }
    }
    throw new Error(msg);
  }
  if (ct.includes("application/json")) {
    return (await res.json()) as T;
  }
  return undefined as T;
}

export type User = {
  id: string;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
  theme?: string;
  avatar_path?: string | null;
  created_at?: string;
  updated_at?: string;
};

/** Mirrors backend `validateUserCreds` / username rune limits (Unicode code points). */
export const REGISTER_USERNAME_MIN = 2;
export const REGISTER_USERNAME_MAX = 64;
export const REGISTER_PASSWORD_MIN = 10;
export const REGISTER_EMAIL_MAX = 254;

export function registerUsernameLengthOk(raw: string): boolean {
  const s = raw.trim();
  const n = [...s].length;
  return n >= REGISTER_USERNAME_MIN && n <= REGISTER_USERNAME_MAX;
}

export function registerPasswordLengthOk(raw: string): boolean {
  return raw.length >= REGISTER_PASSWORD_MIN;
}

/**
 * Client-side email check (backend uses net/mail). Rejects obvious invalid shapes before submit.
 */
export function registerEmailFormatOk(raw: string): boolean {
  const s = raw.trim();
  if (s.length < 5 || s.length > REGISTER_EMAIL_MAX) return false;
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s);
}

import type { User } from './api';

/** Имя для приветствия: логин; если API старый — часть email до @. */
export function userDisplayName(user: User): string {
  const u = user.username?.trim();
  if (u) {
    return u;
  }
  const at = user.email.indexOf('@');
  return at > 0 ? user.email.slice(0, at) : user.email;
}

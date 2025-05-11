export function isUserLoggedIn(): boolean {
  return Boolean(localStorage.getItem('authToken'));
}

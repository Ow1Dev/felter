import { inject } from '@angular/core';
import { type CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const callbackGuard: CanActivateFn = async () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  await authService.waitForInit();

  const params = new URLSearchParams(window.location.search);
  const code = params.get('code');

  if (!code && authService.isAuthenticated()) {
    await router.navigate(['/'], { replaceUrl: true });
    return false;
  }

  return true;
};

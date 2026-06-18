import { inject } from '@angular/core';
import { type CanActivateFn } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { environment } from '../../environments/environment';

export const authGuard: CanActivateFn = async () => {
  const authService = inject(AuthService);

  await authService.waitForInit();

  if (authService.isAuthenticated()) {
    return true;
  }

  window.location.href = `${environment.identityUrl}/login`;
  return false;
};

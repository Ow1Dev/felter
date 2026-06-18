import { Injectable, signal, computed } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Observable, firstValueFrom } from 'rxjs';
import { environment } from '../../environments/environment';

export interface AuthResponse {
  token: string;
  email: string;
}

export interface CurrentUser {
  id: number;
  email: string;
  username: string;
  display_name: string | null;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly TOKEN_KEY = 'felter_auth_token';

  readonly token = signal<string | null>(localStorage.getItem(this.TOKEN_KEY));
  readonly currentUser = signal<CurrentUser | null>(null);
  readonly isAuthenticated = computed(() => this.token() !== null);

  private initResolve!: () => void;
  private readonly initPromise = new Promise<void>(resolve => {
    this.initResolve = resolve;
  });

  constructor(private http: HttpClient) {
    // Defer session restoration so the service is fully constructed
    // before the auth interceptor re-injects it for the /me request.
    queueMicrotask(() => this.restoreSession());
  }

  /** Returns a promise that resolves once auth state has been restored. */
  waitForInit(): Promise<void> {
    return this.initPromise;
  }

  login(): void {
    window.location.href = `${environment.identityUrl}/login`;
  }

  handleCallback(): Observable<AuthResponse> {
    return new Observable(observer => {
      const params = new URLSearchParams(window.location.search);
      const code = params.get('code');

      if (!code) {
        observer.error(new Error('No code in URL'));
        return;
      }

      this.http
        .post<AuthResponse>(`${environment.identityUrl}/callback`, { code })
        .subscribe({
          next: response => {
            this.setToken(response.token);
            this.fetchCurrentUser().then(() => {
              observer.next(response);
              observer.complete();
            });
          },
          error: err => observer.error(err),
        });
    });
  }

  logout(): Observable<void> {
    return new Observable(observer => {
      this.http
        .post<void>(`${environment.identityUrl}/logout`, {})
        .subscribe({
          next: () => {
            this.clearToken();
            observer.next();
            observer.complete();
          },
          error: () => {
            this.clearToken();
            observer.next();
            observer.complete();
          },
        });
    });
  }

  getCurrentUser(): Observable<CurrentUser> {
    return this.http.get<CurrentUser>(`${environment.identityUrl}/me`);
  }

  private restoreSession(): void {
    if (this.isAuthenticated()) {
      this.fetchCurrentUser().finally(() => this.initResolve());
    } else {
      this.initResolve();
    }
  }

  private setToken(token: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
    this.token.set(token);
  }

  private async fetchCurrentUser(): Promise<void> {
    try {
      const user = await firstValueFrom(this.getCurrentUser());
      this.currentUser.set(user);
    } catch (err) {
      if (err instanceof HttpErrorResponse && err.status === 401) {
        this.clearToken();
      }
    }
  }

  private clearToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
    this.token.set(null);
    this.currentUser.set(null);
  }
}
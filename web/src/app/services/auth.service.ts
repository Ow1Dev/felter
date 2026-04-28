import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface AuthResponse {
  token: string;
  email: string;
}

export interface UserInfo {
  sub: string;
  email: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly TOKEN_KEY = 'felter_auth_token';
  private readonly EMAIL_KEY = 'felter_auth_email';

  readonly token = signal<string | null>(localStorage.getItem(this.TOKEN_KEY));
  readonly email = signal<string | null>(localStorage.getItem(this.EMAIL_KEY));
  readonly isAuthenticated = signal<boolean>(this.token() !== null);

  constructor(private http: HttpClient) {}

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
            this.setToken(response.token, response.email);
            window.history.replaceState({}, '', window.location.pathname);
            observer.next(response);
            observer.complete();
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
          error: err => {
            this.clearToken();
            observer.error(err);
          },
        });
    });
  }

  getUserInfo(): Observable<UserInfo> {
    return this.http.get<UserInfo>(`${environment.identityUrl}/userinfo`);
  }

  private setToken(token: string, email: string): void {
    localStorage.setItem(this.TOKEN_KEY, token);
    localStorage.setItem(this.EMAIL_KEY, email);
    this.token.set(token);
    this.email.set(email);
    this.isAuthenticated.set(true);
  }

  private clearToken(): void {
    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.EMAIL_KEY);
    this.token.set(null);
    this.email.set(null);
    this.isAuthenticated.set(false);
  }
}

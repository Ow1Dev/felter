import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login-page',
  imports: [],
  template: `
    <div class="flex h-screen items-center justify-center bg-background text-foreground">
      <div class="flex flex-col items-center gap-4">
        <h1 class="text-2xl font-bold">Welcome to Felter</h1>
        <button
          (click)="login()"
          class="rounded bg-primary px-4 py-2 text-primary-foreground hover:bg-primary/90"
        >
          Sign in with Keycloak
        </button>
      </div>
    </div>
  `,
})
export class LoginPageComponent implements OnInit {
  constructor(
    private authService: AuthService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    if (this.authService.isAuthenticated()) {
      this.router.navigate(['/']);
    }
  }

  login(): void {
    this.authService.login();
  }
}

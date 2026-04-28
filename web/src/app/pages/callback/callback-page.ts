import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-callback-page',
  imports: [],
  template: `
    <div class="flex h-screen items-center justify-center bg-background text-foreground">
      <div class="flex flex-col items-center gap-4">
        @if (error) {
          <div class="text-red-500">Authentication failed. Please try again.</div>
          <button
            (click)="goToLogin()"
            class="rounded bg-primary px-4 py-2 text-primary-foreground hover:bg-primary/90"
          >
            Back to Login
          </button>
        } @else {
          <div>Signing you in...</div>
        }
      </div>
    </div>
  `,
})
export class CallbackPageComponent implements OnInit {
  error = false;

  constructor(
    private authService: AuthService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    this.authService.handleCallback().subscribe({
      next: () => {
        this.router.navigate(['/']);
      },
      error: () => {
        this.error = true;
      },
    });
  }

  goToLogin(): void {
    this.router.navigate(['/login']);
  }
}

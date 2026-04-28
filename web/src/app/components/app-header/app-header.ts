import { Component, inject, OnInit, signal } from '@angular/core';
import { Router } from '@angular/router';
import { NgpAvatar, NgpAvatarFallback } from 'ng-primitives/avatar';
import { NgpMenu, NgpMenuItem, NgpMenuTrigger } from 'ng-primitives/menu';
import { NgpSeparator } from 'ng-primitives/separator';
import { LucideAngularModule } from 'lucide-angular';
import { AuthService, CurrentUser } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';

/** Global app header displayed at the top of the application. */
@Component({
  selector: 'app-app-header',
  imports: [
    NgpAvatar,
    NgpAvatarFallback,
    NgpMenuTrigger,
    NgpMenu,
    NgpMenuItem,
    NgpSeparator,
    LucideAngularModule,
  ],
  styles: [`
    :host {
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 3.5rem; /* 56px */
      padding: 0.75rem 1rem;
      border-bottom: 1px solid var(--sidebar-border);
      background-color: var(--sidebar);
    }

    [ngpMenu] {
      position: fixed;
      z-index: 100;
    }
  `],
  template: `
    <!-- Left side: brand name (clickable) -->
    <button
      (click)="goToWorkspaceChooser()"
      class="text-sm font-semibold text-sidebar-foreground transition-colors hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-ring rounded px-2 py-1"
    >
      felter
    </button>

    <!-- Center: empty for now -->
    <div></div>

    <!-- Right side: user menu -->
    <ng-template #userMenu>
      <div
        ngpMenu
        style="background-color:var(--popover);color:var(--popover-foreground);border:1px solid var(--border);"
        class="min-w-52 rounded-lg p-1 shadow-lg"
      >
        <!-- User info header -->
        <div class="flex items-center gap-2.5 px-2.5 py-2 mb-1">
          <span
            ngpAvatar
            class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-secondary"
          >
            <span ngpAvatarFallback class="text-xs font-semibold text-secondary-foreground">{{ getInitials() }}</span>
          </span>
          <div class="flex flex-col">
            <span class="text-sm font-medium text-popover-foreground">{{ getDisplayName() }}</span>
            <span class="text-xs text-muted-foreground">{{ user()?.email }}</span>
          </div>
        </div>

        <div ngpSeparator style="border-top:1px solid var(--border);" class="my-1"></div>

        <!-- User settings -->
        <button
          ngpMenuItem
          (click)="goToUserSettings()"
          class="menu-item flex w-full items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm outline-none transition-colors cursor-pointer"
        >
          <lucide-icon name="user-round" [size]="14" class="text-muted-foreground" />
          User Settings
        </button>

        <div ngpSeparator style="border-top:1px solid var(--border);" class="my-1"></div>

        <!-- Theme toggle -->
        <button
          ngpMenuItem
          [ngpMenuItemCloseOnSelect]="false"
          (click)="themeService.toggle()"
          class="menu-item flex w-full items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm outline-none transition-colors cursor-pointer"
        >
          @if (themeService.theme() === 'dark') {
            <lucide-icon name="sun" [size]="14" class="text-muted-foreground" />
            Switch to Light Mode
          } @else {
            <lucide-icon name="moon" [size]="14" class="text-muted-foreground" />
            Switch to Dark Mode
          }
        </button>
      </div>
    </ng-template>

    <!-- User avatar trigger -->
    <button
      [ngpMenuTrigger]="userMenu"
      ngpMenuTriggerPlacement="bottom-end"
      class="flex h-9 w-9 items-center justify-center rounded-full bg-secondary transition-colors hover:ring-2 hover:ring-border cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <span class="text-xs font-semibold text-secondary-foreground">{{ getInitials() }}</span>
    </button>
  `,
})
export class AppHeaderComponent implements OnInit {
  protected readonly themeService = inject(ThemeService);
  private readonly router = inject(Router);
  private readonly authService = inject(AuthService);

  protected user = signal<CurrentUser | null>(null);

  ngOnInit(): void {
    if (this.authService.isAuthenticated()) {
      this.authService.getCurrentUser().subscribe({
        next: user => this.user.set(user),
        error: () => this.user.set(null),
      });
    }
  }

  protected goToUserSettings(): void {
    void this.router.navigate(['/settings']);
  }

  protected goToWorkspaceChooser(): void {
    void this.router.navigate(['/']);
  }

  protected getInitials(): string {
    const u = this.user();
    if (!u) return '??';
    const name = u.display_name || u.username || u.email;
    const parts = name.split(/[@\s]/);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  }

  protected getDisplayName(): string {
    const u = this.user();
    if (!u) return 'Unknown';
    return u.display_name || u.username || 'Unknown';
  }
}

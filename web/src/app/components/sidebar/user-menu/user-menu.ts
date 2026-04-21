import { Component, inject, input } from '@angular/core';
import { Router } from '@angular/router';
import { NgpAvatar, NgpAvatarFallback } from 'ng-primitives/avatar';
import { NgpMenu, NgpMenuItem, NgpMenuTrigger } from 'ng-primitives/menu';
import { NgpSeparator } from 'ng-primitives/separator';
import { NgpTooltipTrigger } from 'ng-primitives/tooltip';
import { LucideAngularModule } from 'lucide-angular';
import { ThemeService } from '../../../services/theme.service';
import { WorkspaceRouteService } from '../../../services/workspace-route.service';

/** Bottom user area: avatar, user name, settings & theme toggle menu. */
@Component({
  selector: 'app-user-menu',
  imports: [
    NgpAvatar,
    NgpAvatarFallback,
    NgpMenuTrigger,
    NgpMenu,
    NgpMenuItem,
    NgpSeparator,
    NgpTooltipTrigger,
    LucideAngularModule,
  ],
  styles: [`
    [ngpMenu] {
      position: fixed;
      z-index: 100;
    }
  `],
  template: `
    <!-- User settings menu template -->
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
            <span ngpAvatarFallback class="text-xs font-semibold text-secondary-foreground">JD</span>
          </span>
          <div class="flex flex-col">
            <span class="text-sm font-medium text-popover-foreground">John Doe</span>
            <span class="text-xs text-muted-foreground">john&#64;example.com</span>
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

        <!-- Workspace settings -->
        <button
          ngpMenuItem
          (click)="goToWorkspaceSettings()"
          class="menu-item flex w-full items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm outline-none transition-colors cursor-pointer"
        >
          <lucide-icon name="settings" [size]="14" class="text-muted-foreground" />
          Workspace Settings
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

    <!-- Trigger: expanded -->
    @if (!collapsed()) {
      <button
        [ngpMenuTrigger]="userMenu"
        ngpMenuTriggerPlacement="top-start"
        class="flex w-full items-center gap-2.5 rounded-lg px-2 py-2 text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      >
        <span
          ngpAvatar
          class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-secondary"
        >
          <span ngpAvatarFallback class="text-xs font-semibold text-secondary-foreground">JD</span>
        </span>
        <span class="flex-1 truncate text-left text-sidebar-foreground">John Doe</span>
        <lucide-icon name="ellipsis" [size]="14" class="shrink-0 text-muted-foreground" />
      </button>
    }

    <!-- Trigger: collapsed (avatar only + tooltip) -->
    @if (collapsed()) {
      <button
        [ngpMenuTrigger]="userMenu"
        ngpMenuTriggerPlacement="right-end"
        [ngpTooltipTrigger]="'John Doe'"
        ngpTooltipTriggerPlacement="right"
        [ngpTooltipTriggerShowDelay]="300"
        class="flex h-9 w-9 mx-auto items-center justify-center rounded-full bg-secondary transition-colors hover:ring-2 hover:ring-sidebar-border cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      >
        <span class="text-xs font-semibold text-secondary-foreground">JD</span>
      </button>
    }
  `,
})
export class UserMenuComponent {
  readonly collapsed = input(false);
  protected readonly themeService = inject(ThemeService);
  private readonly router = inject(Router);
  private readonly workspaceRoute = inject(WorkspaceRouteService);

  protected goToUserSettings(): void {
    void this.router.navigate(['/settings']);
  }

  protected goToWorkspaceSettings(): void {
    const slug = this.workspaceRoute.workspaceSlug();
    if (!slug) return;
    void this.router.navigate(['/', slug, 'settings']);
  }
}

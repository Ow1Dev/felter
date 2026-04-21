import { Component, inject, input } from '@angular/core';
import { NgpMenu, NgpMenuItem, NgpMenuTrigger } from 'ng-primitives/menu';
import { NgpTooltipTrigger } from 'ng-primitives/tooltip';
import { LucideAngularModule } from 'lucide-angular';
import { Router } from '@angular/router';
import { WorkspaceService } from '../../../services/workspace.service';
import { WorkspaceRouteService } from '../../../services/workspace-route.service';
import { ViewService } from '../../../services/view.service';

/** Displays the active workspace with a dropdown to switch between workspaces. */
@Component({
  selector: 'app-workspace-switcher',
  imports: [NgpMenuTrigger, NgpMenu, NgpMenuItem, NgpTooltipTrigger, LucideAngularModule],
  styles: [`
    [ngpMenu] {
      position: fixed;
      z-index: 100;
    }
  `],
  template: `
    <!-- Workspace menu template -->
    <ng-template #workspaceMenu>
      <div
        ngpMenu
        style="background-color:var(--popover);color:var(--popover-foreground);border:1px solid var(--border);"
        class="min-w-48 rounded-lg p-1 shadow-lg"
      >
        @for (ws of workspaceService.workspaces(); track ws.id) {
          <button
            ngpMenuItem
            (click)="switchWorkspace(ws.slug)"
            class="menu-item flex w-full items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm outline-none transition-colors cursor-pointer"
          >
            <span
              class="flex h-5 w-5 shrink-0 items-center justify-center rounded text-[10px] font-bold text-white"
              [class]="ws.color"
            >{{ ws.initials }}</span>
            <span class="flex-1 text-left">{{ ws.name }}</span>
            @if (ws.slug === workspaceRoute.workspaceSlug()) {
              <lucide-icon name="check" [size]="14" class="text-sidebar-primary" />
            }
          </button>
        }
      </div>
    </ng-template>

    <!-- Trigger: expanded -->
    @if (!collapsed()) {
      @if (activeWorkspace(); as current) {
        <button
          [ngpMenuTrigger]="workspaceMenu"
          ngpMenuTriggerPlacement="bottom-start"
          class="flex w-full items-center gap-2.5 rounded-lg px-2 py-2 text-sm font-medium text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        >
          <span
            class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-xs font-bold text-white"
            [class]="current.color"
          >{{ current.initials }}</span>
          <span class="flex-1 truncate text-left text-sidebar-foreground">
            {{ current.name }}
          </span>
          <lucide-icon name="chevron-down" [size]="14" class="shrink-0 text-muted-foreground" />
        </button>
      }
    }

    <!-- Trigger: collapsed (icon only + tooltip) -->
    @if (collapsed()) {
      @if (activeWorkspace(); as current) {
        <button
          [ngpMenuTrigger]="workspaceMenu"
          ngpMenuTriggerPlacement="right-start"
          [ngpTooltipTrigger]="current.name"
          ngpTooltipTriggerPlacement="right"
          [ngpTooltipTriggerShowDelay]="300"
          class="flex h-9 w-9 items-center justify-center rounded-lg text-xs font-bold text-white transition-colors hover:opacity-80 cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring mx-auto"
          [class]="current.color"
        >
          {{ current.initials }}
        </button>
      }
    }
  `,
})
export class WorkspaceSwitcherComponent {
  readonly collapsed = input(false);
  private readonly router = inject(Router);
  private readonly viewService = inject(ViewService);
  protected readonly workspaceService = inject(WorkspaceService);
  protected readonly workspaceRoute = inject(WorkspaceRouteService);

  protected readonly activeWorkspace = this.workspaceRoute.workspace;

  protected switchWorkspace(slug: string): void {
    this.workspaceService.setActiveBySlug(slug);

    const urlTree = this.router.parseUrl(this.router.url);
    const primary = urlTree.root.children['primary'];
    const segments = primary?.segments ?? [];

    if (segments.length === 0) {
      this.navigateToDefaultView(slug);
      return;
    }

    const first = segments[0]?.path ?? '';

    if (first === 'settings' || first === 'no-workspace') {
      this.navigateToDefaultView(slug);
      return;
    }

    const rest = segments.slice(1).map(segment => segment.path);
    void this.router.navigate(['/', slug, ...rest], {
      queryParams: urlTree.queryParams,
      fragment: urlTree.fragment ?? undefined,
    });
  }

  private navigateToDefaultView(slug: string): void {
    const defaultView = this.viewService.getDefaultView();
    if (!defaultView) return;
    void this.router.navigate(['/', slug, 'views', defaultView.slug]);
  }
}

import { Component, OnInit, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { NgpTooltipTrigger } from 'ng-primitives/tooltip';
import { LucideAngularModule } from 'lucide-angular';
import { ViewListComponent } from './view-list/view-list';
import { ProjectRouteService } from '../../services/project-route.service';

const COLLAPSED_KEY = 'felter-sidebar-collapsed';

/** Sidebar shell — manages collapsed/expanded state, displays views and project settings. */
@Component({
  selector: 'app-sidebar',
  imports: [
    ViewListComponent,
    NgpTooltipTrigger,
    LucideAngularModule,
  ],
  styles: [`
    :host {
      display: flex;
      flex-direction: column;
      height: 100%;
      flex-shrink: 0;
      overflow: hidden;
      background-color: var(--sidebar);
      border-right: 1px solid var(--sidebar-border);
      transition: width 200ms ease-in-out;
    }

    :host.collapsed {
      width: 3.5rem; /* 56px */
    }

    :host.expanded {
      width: 16rem; /* 256px */
    }
  `],
  host: {
    '[class.collapsed]': 'collapsed()',
    '[class.expanded]': '!collapsed()',
  },
  template: `
    <!-- Top zone: project name + collapse toggle -->
    <div
      class="flex items-center gap-1 border-b border-sidebar-border px-2 py-3"
      [class.justify-center]="collapsed()"
      [class.justify-between]="!collapsed()"
    >
      <!-- Project name: expanded only -->
      @if (!collapsed()) {
        <div class="flex-1 min-w-0">
          <span class="text-sm font-semibold text-sidebar-foreground truncate">
            {{ projectRoute.project()?.name }}
          </span>
        </div>
      }

      <!-- Collapse / expand toggle -->
      <button
        [ngpTooltipTrigger]="collapsed() ? 'Expand sidebar' : 'Collapse sidebar'"
        ngpTooltipTriggerPlacement="right"
        [ngpTooltipTriggerShowDelay]="500"
        (click)="toggleCollapsed()"
        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-sidebar-foreground/40 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        [class.hidden]="collapsed()"
      >
        <lucide-icon name="panel-left-close" [size]="15" />
      </button>

      @if (collapsed()) {
        <button
          [ngpTooltipTrigger]="'Expand sidebar'"
          ngpTooltipTriggerPlacement="right"
          [ngpTooltipTriggerShowDelay]="500"
          (click)="toggleCollapsed()"
          class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-sidebar-foreground/40 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        >
          <lucide-icon name="panel-left-open" [size]="15" />
        </button>
      }
    </div>

    <!-- Middle zone: scrollable view list -->
    <div class="flex-1 overflow-y-auto overflow-x-hidden py-2">
      <app-view-list [collapsed]="collapsed()" />
    </div>

    <!-- Bottom zone: project settings -->
    <div class="border-t border-sidebar-border px-2 py-3">
      <!-- Settings: expanded -->
      @if (!collapsed()) {
        <button
          (click)="goToProjectSettings()"
          class="flex w-full items-center gap-2.5 rounded-lg px-2 py-2 text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        >
          <lucide-icon name="settings" [size]="16" class="shrink-0" />
          <span class="flex-1 text-left text-sidebar-foreground">Settings</span>
        </button>
      }

      <!-- Settings: collapsed (icon only + tooltip) -->
      @if (collapsed()) {
        <button
          [ngpTooltipTrigger]="'Settings'"
          ngpTooltipTriggerPlacement="right"
          [ngpTooltipTriggerShowDelay]="300"
          (click)="goToProjectSettings()"
          class="flex h-9 w-9 mx-auto items-center justify-center rounded-lg transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        >
          <lucide-icon name="settings" [size]="16" class="text-sidebar-foreground" />
        </button>
      }
    </div>
  `,
})
export class SidebarComponent implements OnInit {
  readonly collapsed = signal(false);
  private readonly router = inject(Router);
  protected readonly projectRoute = inject(ProjectRouteService);

  ngOnInit(): void {
    const stored = localStorage.getItem(COLLAPSED_KEY);
    if (stored === 'true') this.collapsed.set(true);
  }

  protected toggleCollapsed(): void {
    this.collapsed.update(v => !v);
    localStorage.setItem(COLLAPSED_KEY, String(this.collapsed()));
  }

  protected goToProjectSettings(): void {
    const slug = this.projectRoute.projectSlug();
    if (!slug) return;
    void this.router.navigate(['/', slug, 'settings']);
  }
}

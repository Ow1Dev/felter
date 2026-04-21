import { Component, inject, input } from '@angular/core';
import { NgpTooltipTrigger } from 'ng-primitives/tooltip';
import { LucideAngularModule } from 'lucide-angular';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { ViewService } from '../../../services/view.service';
import { WorkspaceRouteService } from '../../../services/workspace-route.service';

/** Renders the flat list of data views. Shows icon+label when expanded, icon+tooltip when collapsed. */
@Component({
  selector: 'app-view-list',
  imports: [NgpTooltipTrigger, LucideAngularModule, RouterLink, RouterLinkActive],
  styles: [`
    .view-item {
      color: var(--sidebar-foreground);
      opacity: 0.6;
      border-radius: 0.5rem;
      transition: background-color 150ms ease, opacity 150ms ease;
    }

    .view-item:hover {
      background-color: var(--sidebar-accent);
      color: var(--sidebar-accent-foreground);
      opacity: 1;
    }

    .view-item.active {
      background-color: var(--sidebar-primary);
      color: var(--sidebar-primary-foreground);
      opacity: 1;
      font-weight: 500;
    }

    .view-item.active lucide-icon,
    .view-item.active :host-context(app-view-list) lucide-icon {
      color: var(--sidebar-primary-foreground);
    }

    .view-item:focus-visible {
      outline: 2px solid var(--sidebar-ring);
      outline-offset: -2px;
    }
  `],
  template: `
    @if (workspaceSlug(); as slug) {
      <nav class="flex flex-col gap-0.5 px-2">
        @for (view of viewService.views(); track view.id) {
          @if (collapsed()) {
            <!-- Collapsed: icon only + tooltip -->
            <button
              [ngpTooltipTrigger]="view.name"
              ngpTooltipTriggerPlacement="right"
              [ngpTooltipTriggerShowDelay]="300"
              [routerLink]="['/', slug, 'views', view.slug]"
              routerLinkActive="active"
              class="view-item flex h-9 w-9 items-center justify-center mx-auto"
            >
              <lucide-icon [name]="view.icon" [size]="16" />
            </button>
          } @else {
            <!-- Expanded: icon + name -->
            <button
              [routerLink]="['/', slug, 'views', view.slug]"
              routerLinkActive="active"
              class="view-item flex w-full items-center gap-2.5 px-2.5 py-2 text-sm"
            >
              <lucide-icon [name]="view.icon" [size]="15" class="shrink-0" />
              <span class="flex-1 truncate text-left">{{ view.name }}</span>
            </button>
          }
        }

      </nav>
    } @else {
      <div class="px-4 py-6 text-center text-xs text-muted-foreground">
        Select a workspace to view its boards.
      </div>
    }
  `,
})
export class ViewListComponent {
  readonly collapsed = input(false);
  protected readonly viewService = inject(ViewService);
  protected readonly workspaceSlug = inject(WorkspaceRouteService).workspaceSlug;
}

import { Component, inject, input } from '@angular/core';
import { NgpTooltipTrigger } from 'ng-primitives/tooltip';
import { LucideAngularModule } from 'lucide-angular';
import { ViewService } from '../../../services/view.service';

/** Renders the flat list of data views. Shows icon+label when expanded, icon+tooltip when collapsed. */
@Component({
  selector: 'app-view-list',
  imports: [NgpTooltipTrigger, LucideAngularModule],
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
      background-color: var(--sidebar-foreground);
      color: var(--sidebar);
      opacity: 1;
      font-weight: 500;
    }

    .view-item.active lucide-icon,
    .view-item.active :host-context(app-view-list) lucide-icon {
      color: var(--sidebar);
    }

    .view-item:focus-visible {
      outline: 2px solid var(--sidebar-ring);
      outline-offset: -2px;
    }
  `],
  template: `
    <nav class="flex flex-col gap-0.5 px-2">
      @for (view of viewService.views(); track view.id) {
        @if (collapsed()) {
          <!-- Collapsed: icon only + tooltip -->
          <button
            [ngpTooltipTrigger]="view.name"
            ngpTooltipTriggerPlacement="right"
            [ngpTooltipTriggerShowDelay]="300"
            (click)="viewService.setActive(view.id)"
            class="view-item flex h-9 w-9 items-center justify-center mx-auto cursor-pointer"
            [class.active]="view.id === viewService.activeView()?.id"
          >
            <lucide-icon [name]="view.icon" [size]="16" />
          </button>
        } @else {
          <!-- Expanded: icon + name -->
          <button
            (click)="viewService.setActive(view.id)"
            class="view-item flex w-full items-center gap-2.5 px-2.5 py-2 text-sm cursor-pointer"
            [class.active]="view.id === viewService.activeView()?.id"
          >
            <lucide-icon [name]="view.icon" [size]="15" class="shrink-0" />
            <span class="flex-1 truncate text-left">{{ view.name }}</span>
          </button>
        }
      }
    </nav>
  `,
})
export class ViewListComponent {
  readonly collapsed = input(false);
  protected readonly viewService = inject(ViewService);


}

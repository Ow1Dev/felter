import { Component, OnInit, signal } from '@angular/core';
import { NgpTooltipTrigger } from 'ng-primitives/tooltip';
import { LucideAngularModule } from 'lucide-angular';
import { WorkspaceSwitcherComponent } from './workspace-switcher/workspace-switcher';
import { ViewListComponent } from './view-list/view-list';
import { UserMenuComponent } from './user-menu/user-menu';

const COLLAPSED_KEY = 'felter-sidebar-collapsed';

/** Root sidebar shell — manages collapsed/expanded state, composes child zones. */
@Component({
  selector: 'app-sidebar',
  imports: [
    WorkspaceSwitcherComponent,
    ViewListComponent,
    UserMenuComponent,
    NgpTooltipTrigger,
    LucideAngularModule,
  ],
  styles: [`
    :host {
      display: flex;
      flex-direction: column;
      height: 100vh;
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
    <!-- Top zone: workspace switcher + collapse toggle -->
    <div
      class="flex items-center gap-1 border-b border-sidebar-border px-2 py-3"
      [class.justify-center]="collapsed()"
      [class.justify-between]="!collapsed()"
    >
      @if (!collapsed()) {
        <div class="flex-1 min-w-0">
          <app-workspace-switcher [collapsed]="collapsed()" />
        </div>
      } @else {
        <app-workspace-switcher [collapsed]="collapsed()" />
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

    <!-- Bottom zone: user menu -->
    <div class="border-t border-sidebar-border px-2 py-3">
      <app-user-menu [collapsed]="collapsed()" />
    </div>
  `,
})
export class SidebarComponent implements OnInit {
  readonly collapsed = signal(false);

  ngOnInit(): void {
    const stored = localStorage.getItem(COLLAPSED_KEY);
    if (stored === 'true') this.collapsed.set(true);
  }

  protected toggleCollapsed(): void {
    this.collapsed.update(v => !v);
    localStorage.setItem(COLLAPSED_KEY, String(this.collapsed()));
  }
}

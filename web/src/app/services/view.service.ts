import { Injectable, signal } from '@angular/core';

export type ViewType = 'kanban' | 'list' | 'table' | 'calendar';

export interface AppView {
  id: string;
  name: string;
  /** Lucide icon name to render alongside this view. */
  icon: string;
  viewType: ViewType;
}

const MOCK_VIEWS: AppView[] = [
  { id: 'view-1', name: 'Work Orders', icon: 'briefcase',        viewType: 'kanban'   },
  { id: 'view-2', name: 'Customers',   icon: 'users',            viewType: 'list'     },
  { id: 'view-3', name: 'Assets',      icon: 'wrench',           viewType: 'table'    },
  { id: 'view-4', name: 'Schedule',    icon: 'calendar',         viewType: 'calendar' },
  { id: 'view-5', name: 'Invoices',    icon: 'file-text',        viewType: 'list'     },
  { id: 'view-6', name: 'Technicians', icon: 'hard-hat',         viewType: 'table'    },
];

/** Manages the list of data views and tracks the currently active one. */
@Injectable({ providedIn: 'root' })
export class ViewService {
  readonly views = signal<AppView[]>(MOCK_VIEWS);
  readonly activeView = signal<AppView | null>(MOCK_VIEWS[0]);

  /** Set the active view by id. */
  setActive(id: string): void {
    const view = this.views().find(v => v.id === id);
    if (view) this.activeView.set(view);
  }
}

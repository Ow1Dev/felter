import { Injectable, signal } from '@angular/core';

export type ViewType = 'kanban' | 'list' | 'table' | 'calendar';

export interface AppView {
  id: string;
  slug: string;
  name: string;
  /** Lucide icon name to render alongside this view. */
  icon: string;
  viewType: ViewType;
}

const MOCK_VIEWS: AppView[] = [
  {
    id: 'view-1',
    slug: 'work-orders',
    name: 'Work Orders',
    icon: 'briefcase',
    viewType: 'kanban',
  },
  {
    id: 'view-2',
    slug: 'customers',
    name: 'Customers',
    icon: 'users',
    viewType: 'list',
  },
  {
    id: 'view-3',
    slug: 'assets',
    name: 'Assets',
    icon: 'wrench',
    viewType: 'table',
  },
  {
    id: 'view-4',
    slug: 'schedule',
    name: 'Schedule',
    icon: 'calendar',
    viewType: 'calendar',
  },
  {
    id: 'view-5',
    slug: 'invoices',
    name: 'Invoices',
    icon: 'file-text',
    viewType: 'list',
  },
  {
    id: 'view-6',
    slug: 'technicians',
    name: 'Technicians',
    icon: 'hard-hat',
    viewType: 'table',
  },
];

/** Manages the list of data views and tracks the currently active one. */
@Injectable({ providedIn: 'root' })
export class ViewService {
  readonly views = signal<AppView[]>(MOCK_VIEWS);
  readonly activeView = signal<AppView | null>(MOCK_VIEWS[0]);

  getDefaultView(): AppView | null {
    return this.views()[0] ?? null;
  }

  getBySlug(slug: string): AppView | undefined {
    return this.views().find(v => v.slug === slug);
  }

  /** Set the active view by id. */
  setActive(id: string): void {
    const view = this.views().find(v => v.id === id);
    if (view) this.activeView.set(view);
  }

  /** Set the active view by slug. */
  setActiveBySlug(slug: string): void {
    const view = this.views().find(v => v.slug === slug);
    if (view) this.activeView.set(view);
  }
}

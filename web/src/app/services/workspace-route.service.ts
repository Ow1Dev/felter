import { Injectable, computed, inject, signal } from '@angular/core';
import { ActivatedRouteSnapshot, NavigationEnd, Router } from '@angular/router';
import { filter } from 'rxjs/operators';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { Workspace, WorkspaceService } from './workspace.service';

/** Synchronises the active workspace context with the current router URL. */
@Injectable({ providedIn: 'root' })
export class WorkspaceRouteService {
  private readonly router = inject(Router);
  private readonly workspaceService = inject(WorkspaceService);

  private readonly slug = signal<string | null>(null);

  /** Currently active workspace slug pulled from the router or fallback to active workspace. */
  readonly workspaceSlug = computed(() => {
    const fromRoute = this.slug();
    if (fromRoute) return fromRoute;
    const active = this.workspaceService.activeWorkspace();
    return active?.slug ?? null;
  });

  /** Workspace object for the active slug, or `null` if not found. */
  readonly workspace = computed<Workspace | null>(() => {
    const current = this.workspaceSlug();
    if (!current) return null;
    return this.workspaceService.getBySlug(current) ?? null;
  });

  constructor() {
    this.updateSlugFromRouter();

    this.router.events
      .pipe(filter(event => event instanceof NavigationEnd), takeUntilDestroyed())
      .subscribe(() => this.updateSlugFromRouter());
  }

  private updateSlugFromRouter(): void {
    const snapshot = this.router.routerState.snapshot.root;
    const slug = this.extractSlug(snapshot);
    this.slug.set(slug);

    if (slug) {
      this.workspaceService.setActiveBySlug(slug);
    }
  }

  private extractSlug(route: ActivatedRouteSnapshot | null): string | null {
    if (!route) return null;
    const current = route.paramMap.get('workspaceSlug');
    if (current) return current;
    for (const child of route.children) {
      const slug = this.extractSlug(child);
      if (slug) return slug;
    }
    return null;
  }
}

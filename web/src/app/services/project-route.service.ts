import { Injectable, computed, inject, signal } from '@angular/core';
import { ActivatedRouteSnapshot, NavigationEnd, Router } from '@angular/router';
import { filter } from 'rxjs/operators';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { Project, ProjectService } from './project.service';

/** Synchronises the active project context with the current router URL. */
@Injectable({ providedIn: 'root' })
export class ProjectRouteService {
  private readonly router = inject(Router);
  private readonly projectService = inject(ProjectService);

  private readonly slug = signal<string | null>(null);

  /** Currently active project slug pulled from the router or fallback to active project. */
  readonly projectSlug = computed(() => {
    const fromRoute = this.slug();
    if (fromRoute) return fromRoute;
    const active = this.projectService.activeProject();
    return active?.slug ?? null;
  });

  /** Project object for the active slug, or `null` if not found. */
  readonly project = computed<Project | null>(() => {
    const current = this.projectSlug();
    if (!current) return null;
    return this.projectService.getBySlug(current) ?? null;
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
      this.projectService.setActiveBySlug(slug);
    }
  }

  private extractSlug(route: ActivatedRouteSnapshot | null): string | null {
    if (!route) return null;
    const current = route.paramMap.get('projectSlug');
    if (current) return current;
    for (const child of route.children) {
      const slug = this.extractSlug(child);
      if (slug) return slug;
    }
    return null;
  }
}

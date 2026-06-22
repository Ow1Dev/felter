import { Component, OnInit, effect, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ProjectService } from '../../services/project.service';
import { ViewService } from '../../services/view.service';

/** Redirects `/:projectSlug` to that project's default view, or shows an empty state. */
@Component({
  standalone: true,
  template: `
    @if (isLoading()) {
      <section class="flex flex-1 items-center justify-center p-8 text-muted-foreground">
        <p class="text-sm">Loading project…</p>
      </section>
    } @else {
      <section class="flex flex-1 flex-col items-center justify-center gap-4 text-center p-8">
        <h2 class="text-lg font-semibold text-foreground">No views yet</h2>
        <p class="max-w-md text-sm text-muted-foreground">
          This project doesn't have any views. Views will appear here once they are created.
        </p>
      </section>
    }
  `,
})
export class ProjectRedirectComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly projectService = inject(ProjectService);
  private readonly viewService = inject(ViewService);

  protected readonly isLoading = signal(true);
  private readonly handled = signal(false);

  constructor() {
    effect(() => {
      if (this.handled()) return;
      const slug = this.route.snapshot.paramMap.get('projectSlug');
      if (!slug || this.projectService.projects().length === 0) return;

      this.handled.set(true);
      const project = this.projectService.getBySlug(slug);
      if (!project) {
        void this.router.navigate(['/no-project']);
        return;
      }

      const defaultView = this.viewService.getDefaultView();
      if (defaultView) {
        void this.router.navigate(['/', project.slug, 'views', defaultView.slug]);
        return;
      }

      // No views — stay inside the project layout and show the empty state
      this.isLoading.set(false);
    });
  }

  ngOnInit(): void {
    const slug = this.route.snapshot.paramMap.get('projectSlug');
    if (!slug) {
      void this.router.navigate(['/no-project']);
      return;
    }

    if (this.projectService.projects().length === 0) {
      this.projectService.loadProject(slug);
    }
  }
}

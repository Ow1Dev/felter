import { Component, OnInit, inject } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ProjectService } from '../../services/project.service';
import { ViewService } from '../../services/view.service';

/** Redirects `/:projectSlug` to that project's default view, validating the slug. */
@Component({
  standalone: true,
  template: `
    <section class="flex flex-1 items-center justify-center p-8 text-muted-foreground">
      <p class="text-sm">Loading project…</p>
    </section>
  `,
})
export class ProjectRedirectComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly projectService = inject(ProjectService);
  private readonly viewService = inject(ViewService);

  ngOnInit(): void {
    const slug = this.route.snapshot.paramMap.get('projectSlug');
    if (!slug) {
      void this.router.navigate(['/no-project']);
      return;
    }

    const project = this.projectService.getBySlug(slug);
    if (!project) {
      void this.router.navigate(['/no-project']);
      return;
    }

    const defaultView = this.viewService.getDefaultView();
    if (!defaultView) {
      void this.router.navigate(['/no-project']);
      return;
    }

    void this.router.navigate(['/', project.slug, 'views', defaultView.slug]);
  }
}

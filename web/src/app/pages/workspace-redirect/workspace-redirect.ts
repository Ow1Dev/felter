import { Component, OnInit, inject } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { WorkspaceService } from '../../services/workspace.service';
import { ViewService } from '../../services/view.service';

/** Redirects `/:workspaceSlug` to that workspace's default view, validating the slug. */
@Component({
  standalone: true,
  template: `
    <section class="flex flex-1 items-center justify-center p-8 text-muted-foreground">
      <p class="text-sm">Loading workspace…</p>
    </section>
  `,
})
export class WorkspaceRedirectComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly workspaceService = inject(WorkspaceService);
  private readonly viewService = inject(ViewService);

  ngOnInit(): void {
    const slug = this.route.snapshot.paramMap.get('workspaceSlug');
    if (!slug) {
      void this.router.navigate(['/no-workspace']);
      return;
    }

    const workspace = this.workspaceService.workspaces().find(ws => ws.slug === slug);
    if (!workspace) {
      void this.router.navigate(['/no-workspace']);
      return;
    }

    const defaultView = this.viewService.getDefaultView();
    if (!defaultView) {
      void this.router.navigate(['/no-workspace']);
      return;
    }

    void this.router.navigate(['/', workspace.slug, 'views', defaultView.slug]);
  }
}

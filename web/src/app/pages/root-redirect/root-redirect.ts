import { Component, OnInit, inject } from '@angular/core';
import { Router } from '@angular/router';
import { WorkspaceService } from '../../services/workspace.service';
import { ViewService } from '../../services/view.service';

/** Redirects the bare root URL to the active workspace's default view (or fallback page). */
@Component({
  standalone: true,
  template: `
    <section class="flex flex-1 items-center justify-center p-8 text-muted-foreground">
      <p class="text-sm">Preparing your workspace…</p>
    </section>
  `,
})
export class RootRedirectComponent implements OnInit {
  private readonly router = inject(Router);
  private readonly workspaceService = inject(WorkspaceService);
  private readonly viewService = inject(ViewService);

  ngOnInit(): void {
    queueMicrotask(() => {
      let workspace = this.workspaceService.getDefaultWorkspace();
      const storedSlug = this.workspaceService.getStoredActiveWorkspaceSlug();
      if (storedSlug) {
        const stored = this.workspaceService.getBySlug(storedSlug);
        if (stored) workspace = stored;
      }

      if (!workspace) {
        void this.router.navigate(['/no-workspace']);
        return;
      }

      const defaultView = this.viewService.getDefaultView();
      if (!defaultView) {
        void this.router.navigate(['/no-workspace']);
        return;
      }

      this.workspaceService.setActiveBySlug(workspace.slug);
      void this.router.navigate(['/', workspace.slug, 'views', defaultView.slug]);
    });
  }
}

import { Component } from '@angular/core';

/** Placeholder screen displayed when no workspaces are available. */
@Component({
  standalone: true,
  template: `
    <section class="flex flex-1 flex-col items-center justify-center gap-4 p-12 text-center">
      <h1 class="text-2xl font-semibold text-foreground">No workspace yet</h1>
      <p class="max-w-md text-sm text-muted-foreground">
        We couldn’t find any workspaces. Create one in the backend and reload the
        app to get started.
      </p>
      <button
        type="button"
        class="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground/80 hover:text-foreground"
        (click)="reload()"
      >
        Refresh
      </button>
    </section>
  `,
})
export class NoWorkspacePageComponent {
  reload(): void {
    window.location.reload();
  }
}

import { Component } from '@angular/core';

/** Placeholder page for user-level settings. */
@Component({
  standalone: true,
  template: `
    <section class="flex flex-1 flex-col gap-6 overflow-auto p-10">
      <header class="flex flex-col gap-2">
        <h1 class="text-3xl font-semibold tracking-tight text-foreground">
          User settings
        </h1>
        <p class="text-sm text-muted-foreground">
          Personal preferences will live here soon. For now this is just a mock
          screen.
        </p>
      </header>
      <article class="rounded-lg border border-border bg-card p-6 text-sm text-muted-foreground">
        <p>
          Add user-specific controls (profile, notifications, theme) when
          requirements are defined.
        </p>
      </article>
    </section>
  `,
})
export class UserSettingsPageComponent {}

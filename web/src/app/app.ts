import { Component } from '@angular/core';
import { SidebarComponent } from './components/sidebar/sidebar';

/** Root application shell: sidebar + main content area. */
@Component({
  selector: 'app-root',
  imports: [SidebarComponent],
  template: `
    <div class="flex h-screen overflow-hidden bg-background text-foreground">
      <app-sidebar />
      <main class="flex flex-1 flex-col items-center justify-center overflow-auto p-8">
        <p class="text-sm text-muted-foreground">
          Select a view from the sidebar to get started.
        </p>
      </main>
    </div>
  `,
})
export class App {}

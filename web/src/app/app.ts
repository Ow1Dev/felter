import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from './components/sidebar/sidebar';

/** Root application shell: sidebar + main content area. */
@Component({
  selector: 'app-root',
  imports: [SidebarComponent, RouterOutlet],
  template: `
    <div class="flex h-screen overflow-hidden bg-background text-foreground">
      <app-sidebar />
      <main class="flex flex-1 flex-col overflow-auto">
        <router-outlet />
      </main>
    </div>
  `,
})
export class App {}

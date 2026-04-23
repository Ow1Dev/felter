import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { AppHeaderComponent } from './components/app-header/app-header';

/** Root application shell: header + main content area. */
@Component({
  selector: 'app-root',
  imports: [AppHeaderComponent, RouterOutlet],
  template: `
    <div class="flex h-screen flex-col overflow-hidden bg-background text-foreground">
      <!-- Global app header -->
      <app-app-header />

      <!-- Main content area (layout determined by router) -->
      <div class="flex flex-1 w-full">
        <router-outlet />
      </div>
    </div>
  `,
})
export class App {}

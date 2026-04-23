import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from '../../components/sidebar/sidebar';

/** Layout for all workspace-related pages: sidebar + main content area. */
@Component({
  selector: 'app-workspace-layout',
  imports: [SidebarComponent, RouterOutlet],
  template: `
    <div class="flex w-full h-full overflow-hidden">
      <app-sidebar />
      <main class="flex flex-1 flex-col overflow-auto w-full">
        <router-outlet />
      </main>
    </div>
  `,
})
export class WorkspaceLayoutComponent {}

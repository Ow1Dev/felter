import { Component, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { TextInputComponent } from '../../components/ui/text-input/text-input';
import { WorkspaceService } from '../../services/workspace.service';
import { ViewService } from '../../services/view.service';

/** Page for selecting a workspace. */
@Component({
  selector: 'app-workspace-selector',
  standalone: true,
  imports: [TextInputComponent, LucideAngularModule],
  styles: [`
    :host {
      display: flex;
      width: 100%;
      height: 100%;
    }
  `],
  template: `
    <section class="flex flex-col w-full h-full overflow-hidden">
      <!-- Header with search -->
      <div class="flex items-center justify-between gap-4 border-b border-sidebar-border px-8 py-6 w-full" style="background-color: var(--sidebar-accent);">
        <!-- Left side: empty -->
        <!-- Left side: empty -->
        <div></div>

         <!-- Right side: search -->
         <div class="relative w-full max-w-xs">
           <lucide-icon
             name="search"
             [size]="16"
             class="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
           />
           <app-text-input
             placeholder="Search workspaces..."
             [(value)]="searchQuery"
             class="[&_input]:pl-14"
             style="background-color: var(--background);"
           />
         </div>
      </div>

      <!-- Workspace grid -->
      <div class="flex-1 overflow-auto p-8 w-full">
        @if (filteredWorkspaces().length > 0) {
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 w-full">
            @for (workspace of filteredWorkspaces(); track workspace.id) {
              <button
                (click)="selectWorkspace(workspace.slug)"
                class="flex flex-col items-center gap-3 rounded-lg p-6 transition-colors cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-ring"
                style="background-color: var(--sidebar); border: 1px solid var(--sidebar-border);"
                onmouseover="this.style.backgroundColor = 'var(--sidebar-accent)'"
                onmouseout="this.style.backgroundColor = 'var(--sidebar)'"
              >
                <div
                  class="flex h-12 w-12 items-center justify-center rounded-md text-lg font-bold text-white"
                  [class]="workspace.color"
                >
                  {{ workspace.initials }}
                </div>
                <div class="flex flex-col gap-1">
                  <span class="font-semibold text-foreground">{{ workspace.name }}</span>
                  <span class="text-xs text-muted-foreground">{{ workspace.slug }}</span>
                </div>
              </button>
            }
          </div>
        } @else if (searchQuery()) {
          <div class="flex flex-1 flex-col items-center justify-center gap-4 text-center">
            <lucide-icon name="search" [size]="48" class="text-muted-foreground/40" />
            <p class="text-sm text-muted-foreground">No workspaces found matching "{{ searchQuery() }}"</p>
          </div>
        } @else {
          <div class="flex flex-col gap-4 text-center">
            <p class="max-w-md text-sm text-muted-foreground">
              We couldn't find any workspaces. Create one in the backend and reload the app to get started.
            </p>
            <button
              type="button"
              class="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground/80 hover:text-foreground"
              (click)="reload()"
            >
              Refresh
            </button>
          </div>
        }
      </div>
    </section>
  `,
})
export class WorkspaceSelectorComponent {
  protected readonly workspaceService = inject(WorkspaceService);
  protected readonly searchQuery = signal('');
  private readonly router = inject(Router);
  private readonly viewService = inject(ViewService);

  protected readonly filteredWorkspaces = computed(() => {
    return this.workspaceService.searchWorkspaces(this.searchQuery());
  });

  protected selectWorkspace(slug: string): void {
    this.workspaceService.setActiveBySlug(slug);
    const defaultView = this.viewService.getDefaultView();
    if (defaultView) {
      void this.router.navigate(['/', slug, 'views', defaultView.slug]);
    } else {
      void this.router.navigate(['/', slug]);
    }
  }

  protected reload(): void {
    window.location.reload();
  }
}

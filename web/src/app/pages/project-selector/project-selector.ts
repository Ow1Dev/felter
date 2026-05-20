import { Component, computed, inject, signal, viewChild } from '@angular/core';
import { Router } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { TextInputComponent } from '../../components/ui/text-input/text-input';
import { CreateProjectDrawerComponent } from '../../components/create-project-drawer/create-project-drawer';
import { ProjectService } from '../../services/project.service';
import { ViewService } from '../../services/view.service';

/** Page for selecting a project. */
@Component({
  selector: 'app-project-selector',
  standalone: true,
  imports: [TextInputComponent, LucideAngularModule, CreateProjectDrawerComponent],
  styles: [`
    :host {
      display: flex;
      width: 100%;
      height: 100%;
    }
  `],
  template: `
    <section class="flex flex-col w-full h-full overflow-hidden">
      <!-- Header with search + new project -->
      <div class="flex items-center justify-between gap-4 border-b border-sidebar-border px-8 py-6 w-full" style="background-color: var(--sidebar-accent);">
        <div></div>

        <div class="flex items-center gap-3">
          <button
            type="button"
            class="flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            (click)="openDrawer()"
          >
            <lucide-icon name="plus" [size]="16" />
            New Project
          </button>

          <div class="relative w-full max-w-xs">
            <lucide-icon
              name="search"
              [size]="16"
              class="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
            />
            <app-text-input
              placeholder="Search projects..."
              [(value)]="searchQuery"
              class="[&_input]:pl-14"
              style="background-color: var(--background);"
            />
          </div>
        </div>
      </div>

      <!-- Project grid -->
      <div class="flex-1 overflow-auto p-8 w-full">
        @if (filteredProjects().length > 0) {
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 w-full">
            @for (display of filteredProjects(); track display.project.id) {
              <button
                (click)="selectProject(display.project.slug)"
                class="flex flex-col items-center gap-3 rounded-lg p-6 transition-colors cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-ring"
                style="background-color: var(--sidebar); border: 1px solid var(--sidebar-border);"
                onmouseover="this.style.backgroundColor = 'var(--sidebar-accent)'"
                onmouseout="this.style.backgroundColor = 'var(--sidebar)'"
              >
                <div
                  class="flex h-12 w-12 items-center justify-center rounded-md text-lg font-bold text-white"
                  [class]="display.color"
                >
                  {{ display.initials }}
                </div>
                <div class="flex flex-col gap-1">
                  <span class="font-semibold text-foreground">{{ display.project.name }}</span>
                  <span class="text-xs text-muted-foreground">{{ display.project.slug }}</span>
                </div>
              </button>
            }
          </div>
        } @else if (searchQuery()) {
          <div class="flex flex-1 flex-col items-center justify-center gap-4 text-center">
            <lucide-icon name="search" [size]="48" class="text-muted-foreground/40" />
            <p class="text-sm text-muted-foreground">No projects found matching "{{ searchQuery() }}"</p>
          </div>
        } @else {
          <div class="flex flex-col gap-4 text-center">
            <p class="max-w-md text-sm text-muted-foreground">
              We couldn't find any projects. Create one using the New Project button to get started.
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

    <app-create-project-drawer #drawer (create)="onCreateProject($event)" />
  `,
})
export class ProjectSelectorComponent {
  protected readonly searchQuery = signal('');

  constructor() {
    this.projectService.loadProjects();
  }
  private readonly router = inject(Router);
  private readonly viewService = inject(ViewService);
  private readonly projectService = inject(ProjectService);

  private readonly drawerRef = viewChild.required<CreateProjectDrawerComponent>('drawer');

  protected readonly filteredProjects = computed(() => {
    const projects = this.projectService.searchProjects(this.searchQuery());
    return projects.map(p => this.projectService.toDisplay(p));
  });

  protected selectProject(slug: string): void {
    const project = this.projectService.getBySlug(slug);
    if (project) {
      this.projectService.setActive(project);
    }
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

  protected openDrawer(): void {
    this.drawerRef().open();
  }

  protected onCreateProject(event: { name: string; description: string }): void {
    const req = {
      name: event.name,
      description: event.description || undefined,
    };
    this.projectService.createProject(req).subscribe({
      next: project => {
        this.projectService.loadProjects();
        void this.router.navigate(['/projects', project.slug]);
      },
      error: err => {
        console.error('Failed to create project:', err);
      },
    });
  }
}

import { NgClass } from '@angular/common';
import { Component, computed, effect, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { map } from 'rxjs/operators';
import { toSignal } from '@angular/core/rxjs-interop';

import { FormFieldComponent } from '../../components/ui/form-field/form-field';
import { TextInputComponent } from '../../components/ui/text-input/text-input';
import { ProjectService } from '../../services/project.service';
import { ProjectRouteService } from '../../services/project-route.service';

interface SettingsTab {
  id: 'general' | 'datafields' | 'view';
  label: string;
}

const TABS: SettingsTab[] = [
  { id: 'general', label: 'General' },
  { id: 'datafields', label: 'Data' },
  { id: 'view', label: 'View' },
];

@Component({
  standalone: true,
  imports: [NgClass, FormFieldComponent, TextInputComponent],
  host: {
    class: 'flex flex-1 min-h-full',
  },
  styles: [
    `
      .nav-shell {
        background-color: var(--sidebar);
        color: var(--sidebar-foreground);
      }

      .tab-item {
        color: color-mix(in srgb, var(--sidebar-foreground) 70%, transparent);
        border-radius: 0.75rem;
        transition: background-color 150ms ease, color 150ms ease;
      }

      .tab-item:hover,
      .tab-item:focus-visible {
        background-color: var(--sidebar-accent);
        color: var(--sidebar-accent-foreground);
        outline: none;
      }

      .tab-item.active {
        background-color: var(--sidebar-primary);
        color: var(--sidebar-primary-foreground);
        box-shadow: 0 10px 25px -15px color-mix(in srgb, var(--sidebar-primary) 60%, transparent);
      }
    `,
  ],
  template: `
    <section class="flex flex-1 flex-col overflow-hidden">
      <div class="flex flex-1 flex-col overflow-hidden lg:flex-row">
        <nav class="nav-shell flex shrink-0 flex-col border-b border-border px-4 py-6 lg:w-64 lg:border-r lg:border-b-0 lg:overflow-y-auto">
          <h2 class="hidden px-2 text-xs font-semibold uppercase tracking-[0.3em] text-sidebar-foreground/60 lg:block">Project</h2>
          <ul class="mt-4 flex flex-row gap-2 lg:flex-col lg:gap-1.5">
            @for (tab of tabs; track tab.id) {
              <li>
                <button
                  type="button"
                  (click)="selectTab(tab.id)"
                  class="tab-item flex w-full items-center gap-2 px-4 py-3 text-sm font-medium"
                  [ngClass]="{ active: activeTab() === tab.id }"
                >
                  <span class="block font-medium">{{ tab.label }}</span>
                </button>
              </li>
            }
          </ul>
        </nav>

        <div class="flex flex-1 flex-col gap-10 overflow-y-auto px-6 py-10 lg:px-10">
          @if (project(); as project) {
            @switch (activeTab()) {
              @case ('general') {
                <section class="flex flex-col gap-6">
                  <div class="max-w-xl rounded-xl bg-card px-6 py-6">
                    <h2 data-testid="project-title-heading" class="text-lg font-semibold text-foreground">Project title</h2>
                    <p class="mt-1 text-sm text-muted-foreground">
                      Update how this project appears across navigation and integrations.
                    </p>
                    <div class="mt-5 flex flex-col gap-4">
                      <app-form-field
                        [label]="'Project name'"
                        [description]="'Visible to everyone in the project.'"
                        [for]="nameInputId"
                      >
                        <app-text-input
                          [id]="nameInputId"
                          [value]="projectName()"
                          (valueChange)="onNameInput($event)"
                          (blur)="commitName()"
                        />
                      </app-form-field>

                      <div class="rounded-md bg-muted px-3 py-2 text-xs text-muted-foreground">
                        <span class="font-medium text-foreground">Project slug:</span>
                        <code class="ml-1 rounded bg-muted-foreground/10 px-1 py-0.5 text-[11px] text-muted-foreground">
                          {{ project.slug }}
                        </code>
                      </div>
                    </div>
                  </div>
                </section>
              }
              @case ('datafields') {
                <section class="flex flex-col gap-3">
                  <h2 class="text-lg font-semibold text-foreground">Data settings</h2>
                  <p class="text-sm text-muted-foreground">
                    Data settings for {{ project.name }} are coming soon.
                  </p>
                </section>
              }
              @case ('view') {
                <section class="flex flex-col gap-3">
                  <h2 class="text-lg font-semibold text-foreground">View settings</h2>
                  <p class="text-sm text-muted-foreground">
                    Customize how views appear for {{ project.name }} — feature under construction.
                  </p>
                </section>
              }
            }
          } @else {
            <section class="flex flex-1 flex-col items-center justify-center gap-2 text-center">
              <h2 class="text-lg font-semibold text-foreground">Project not found</h2>
              <p class="text-sm text-muted-foreground">
                The requested project could not be located. Switch to another project from the sidebar.
              </p>
            </section>
          }
        </div>
      </div>
    </section>
  `,
})
export class ProjectSettingsPageComponent {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly projectService = inject(ProjectService);
  private readonly projectRoute = inject(ProjectRouteService);

  protected readonly tabs = TABS;
  protected readonly project = this.projectRoute.project;
  protected readonly projectSlug = this.projectRoute.projectSlug;

  private readonly tabParam = toSignal(
    this.route.paramMap.pipe(map(params => params.get('tab'))),
    { initialValue: this.route.snapshot.paramMap.get('tab') },
  );

  protected readonly activeTab = computed(() => {
    const tab = this.tabParam();
    if (!tab) return 'general';
    return TABS.some(t => t.id === tab) ? tab : 'general';
  });

  protected readonly nameInputId = 'project-name-input';
  protected readonly projectName = signal('');

  constructor() {
    effect(
      () => {
        const current = this.project();
        if (current) this.projectName.set(current.name);
      },
      { allowSignalWrites: true },
    );
  }

  protected selectTab(tabId: SettingsTab['id']): void {
    const slug = this.projectSlug();
    if (!slug) return;

    if (tabId === 'general') {
      void this.router.navigate(['/', slug, 'settings'], {
        queryParamsHandling: 'preserve',
      });
    } else {
      void this.router.navigate(['/', slug, 'settings', tabId], {
        queryParamsHandling: 'preserve',
      });
    }
  }

  protected onNameInput(value: string): void {
    this.projectName.set(value);
  }

  protected commitName(): void {
    const project = this.project();
    const slug = this.projectSlug();
    const trimmed = this.projectName().trim();

    if (!project || !slug) return;
    if (!trimmed || trimmed === project.name) {
      this.projectName.set(project.name);
      return;
    }

    // TODO: wire to backend update when available
    console.warn('Project rename not yet implemented');
  }
}

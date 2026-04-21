import { NgClass } from '@angular/common';
import { Component, computed, effect, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { map } from 'rxjs/operators';
import { toSignal } from '@angular/core/rxjs-interop';

import { FormFieldComponent } from '../../components/ui/form-field/form-field';
import { TextInputComponent } from '../../components/ui/text-input/text-input';
import { WorkspaceService } from '../../services/workspace.service';
import { WorkspaceRouteService } from '../../services/workspace-route.service';

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
          <h2 class="hidden px-2 text-xs font-semibold uppercase tracking-[0.3em] text-sidebar-foreground/60 lg:block">Workspace</h2>
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
          @if (workspace(); as workspace) {
            @switch (activeTab()) {
              @case ('general') {
                <section class="flex flex-col gap-6">
                  <div class="max-w-xl rounded-xl bg-card px-6 py-6">
                    <h2 data-testid="workspace-title-heading" class="text-lg font-semibold text-foreground">Workspace title</h2>
                    <p class="mt-1 text-sm text-muted-foreground">
                      Update how this workspace appears across navigation, emails, and integrations.
                    </p>
                    <div class="mt-5 flex flex-col gap-4">
                      <app-form-field
                        [label]="'Workspace name'"
                        [description]="'Visible to everyone in the workspace.'"
                        [for]="nameInputId"
                      >
                        <app-text-input
                          [id]="nameInputId"
                          [value]="workspaceName()"
                          (valueChange)="onNameInput($event)"
                          (blur)="commitName()"
                        />
                      </app-form-field>

                      <div class="rounded-md bg-muted px-3 py-2 text-xs text-muted-foreground">
                        <span class="font-medium text-foreground">Workspace slug:</span>
                        <code class="ml-1 rounded bg-muted-foreground/10 px-1 py-0.5 text-[11px] text-muted-foreground">
                          {{ workspace.slug }}
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
                    Data settings for {{ workspace.name }} are coming soon.
                  </p>
                </section>
              }
              @case ('view') {
                <section class="flex flex-col gap-3">
                  <h2 class="text-lg font-semibold text-foreground">View settings</h2>
                  <p class="text-sm text-muted-foreground">
                    Customize how views appear for {{ workspace.name }} — feature under construction.
                  </p>
                </section>
              }
            }
          } @else {
            <section class="flex flex-1 flex-col items-center justify-center gap-2 text-center">
              <h2 class="text-lg font-semibold text-foreground">Workspace not found</h2>
              <p class="text-sm text-muted-foreground">
                The requested workspace could not be located. Switch to another workspace from the sidebar.
              </p>
            </section>
          }
        </div>
      </div>
    </section>
  `,
})
export class WorkspaceSettingsPageComponent {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly workspaceService = inject(WorkspaceService);
  private readonly workspaceRoute = inject(WorkspaceRouteService);

  protected readonly tabs = TABS;
  protected readonly workspace = this.workspaceRoute.workspace;
  protected readonly workspaceSlug = this.workspaceRoute.workspaceSlug;

  private readonly tabParam = toSignal(
    this.route.paramMap.pipe(map(params => params.get('tab'))),
    { initialValue: this.route.snapshot.paramMap.get('tab') },
  );

  protected readonly activeTab = computed(() => {
    const tab = this.tabParam();
    if (!tab) return 'general';
    return TABS.some(t => t.id === tab) ? tab : 'general';
  });

  protected readonly nameInputId = 'workspace-name-input';
  protected readonly workspaceName = signal('');

  constructor() {
    effect(
      () => {
        const current = this.workspace();
        if (current) this.workspaceName.set(current.name);
      },
      { allowSignalWrites: true },
    );
  }

  protected selectTab(tabId: SettingsTab['id']): void {
    const slug = this.workspaceSlug();
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
    this.workspaceName.set(value);
  }

  protected commitName(): void {
    const workspace = this.workspace();
    const slug = this.workspaceSlug();
    const trimmed = this.workspaceName().trim();

    if (!workspace || !slug) return;
    if (!trimmed || trimmed === workspace.name) {
      this.workspaceName.set(workspace.name);
      return;
    }

    this.workspaceService.updateWorkspaceName(slug, trimmed);
  }
}

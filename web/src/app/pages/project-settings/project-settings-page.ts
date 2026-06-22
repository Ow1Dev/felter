import { NgClass } from '@angular/common';
import { Component, computed, effect, inject, model, signal } from '@angular/core';
import { ActivatedRoute, NavigationEnd, Router } from '@angular/router';
import { filter, map, startWith } from 'rxjs/operators';
import { toSignal } from '@angular/core/rxjs-interop';

import { FormFieldComponent } from '../../components/ui/form-field/form-field';
import { TextInputComponent } from '../../components/ui/text-input/text-input';
import { SelectInputComponent } from '../../components/ui/select-input/select-input';
import { ModalComponent } from '../../components/ui/modal/modal';
import { FieldService, FIELD_TYPES, type FieldType } from '../../services/field.service';
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
  imports: [NgClass, FormFieldComponent, TextInputComponent, SelectInputComponent, ModalComponent],
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
                @if (activeSchemaKey(); as schemaKey) {
                  <!-- Schema Detail View -->
                  <section class="flex flex-col gap-6">
                    <div class="flex items-center justify-between">
                      <button
                        type="button"
                        (click)="goToSchemaList()"
                        class="text-sm text-muted-foreground hover:text-foreground transition-colors"
                      >
                        &larr; Back to schemas
                      </button>
                      <button
                        type="button"
                        (click)="openFieldModal()"
                        class="inline-flex items-center justify-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90"
                      >
                        + Add field
                      </button>
                    </div>

                    @if (activeSchema(); as schema) {
                      <div class="flex flex-col gap-2">
                        <h2 class="text-lg font-semibold text-foreground">{{ schema.name }}</h2>
                        <p class="text-sm text-muted-foreground">Key: <code class="text-xs bg-muted px-1 py-0.5 rounded">{{ schema.key }}</code></p>
                      </div>

                      <div class="w-full">
                        @if (schema.fields && schema.fields.length > 0) {
                          <div class="rounded-xl border border-border overflow-hidden">
                            <table class="w-full text-sm">
                              <thead class="bg-muted">
                                <tr>
                                  <th class="px-4 py-2 text-left font-medium text-muted-foreground">Key</th>
                                  <th class="px-4 py-2 text-left font-medium text-muted-foreground">Type</th>
                                  <th class="px-4 py-2 text-right font-medium text-muted-foreground"></th>
                                </tr>
                              </thead>
                              <tbody>
                                @for (field of schema.fields; track field.key) {
                                  <tr class="border-t border-border">
                                    <td class="px-4 py-3 font-medium text-foreground">{{ field.key }}</td>
                                    <td class="px-4 py-3 text-muted-foreground">
                                      <code class="text-xs bg-muted px-1 py-0.5 rounded">{{ field.type }}</code>
                                    </td>
                                    <td class="px-4 py-3 text-right">
                                      <button
                                        type="button"
                                        (click)="removeField(field.key)"
                                        class="text-xs text-destructive hover:underline"
                                      >
                                        Remove
                                      </button>
                                    </td>
                                  </tr>
                                }
                              </tbody>
                            </table>
                          </div>
                        } @else {
                          <p class="text-sm text-muted-foreground">No fields defined yet. Click "+ Add field" to create one.</p>
                        }
                      </div>
                    } @else {
                      <p class="text-sm text-muted-foreground">Loading schema...</p>
                    }
                  </section>
                } @else {
                  <!-- Schema List View -->
                  <section class="flex flex-col gap-6">
                    <div class="flex items-center justify-between">
                      <div class="flex flex-col gap-2">
                        <h2 class="text-lg font-semibold text-foreground">Schemas</h2>
                        <p class="text-sm text-muted-foreground">
                          Define data schemas for {{ project.name }}. Each schema holds a collection of typed fields.
                        </p>
                      </div>
                      <button
                        type="button"
                        (click)="openSchemaModal()"
                        class="inline-flex items-center justify-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90"
                      >
                        + New schema
                      </button>
                    </div>

                    <div class="w-full">
                      @if (schemas().length > 0) {
                        <div class="rounded-xl border border-border overflow-hidden">
                          <table class="w-full text-sm">
                            <thead class="bg-muted">
                              <tr>
                                <th class="px-4 py-2 text-left font-medium text-muted-foreground">Name</th>
                                <th class="px-4 py-2 text-left font-medium text-muted-foreground">Key</th>
                                <th class="px-4 py-2 text-right font-medium text-muted-foreground"></th>
                              </tr>
                            </thead>
                            <tbody>
                              @for (schema of schemas(); track schema.key) {
                                <tr class="border-t border-border">
                                  <td class="px-4 py-3 font-medium text-foreground">{{ schema.name }}</td>
                                  <td class="px-4 py-3 text-muted-foreground">
                                    <code class="text-xs bg-muted px-1 py-0.5 rounded">{{ schema.key }}</code>
                                  </td>
                                  <td class="px-4 py-3 text-right">
                                    <div class="flex items-center justify-end gap-3">
                                      <button
                                        type="button"
                                        (click)="goToSchema(schema.key)"
                                        class="text-sm text-primary hover:underline"
                                      >
                                        Edit fields
                                      </button>
                                      <button
                                        type="button"
                                        (click)="deleteSchema(schema.key)"
                                        class="text-sm text-destructive hover:underline"
                                      >
                                        Delete
                                      </button>
                                    </div>
                                  </td>
                                </tr>
                              }
                            </tbody>
                          </table>
                        </div>
                      } @else {
                        <p class="text-sm text-muted-foreground">No schemas yet. Click "+ New schema" to create one.</p>
                      }
                    </div>
                  </section>
                }
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

    <!-- Modals rendered at the bottom of the DOM -->
    <app-modal [(isOpen)]="showSchemaModal" title="Create schema" (backdropClick)="closeSchemaModal()">
      <div class="flex flex-col gap-4">
        <app-form-field [label]="'Schema key'" [for]="'schema-key-input'">
          <app-text-input
            id="schema-key-input"
            [value]="modalSchemaKey()"
            (valueChange)="modalSchemaKey.set($event)"
            placeholder="e.g. tasks, orders, customers"
          />
        </app-form-field>

        <app-form-field [label]="'Schema name'" [for]="'schema-name-input'">
          <app-text-input
            id="schema-name-input"
            [value]="modalSchemaName()"
            (valueChange)="modalSchemaName.set($event)"
            placeholder="e.g. Tasks, Orders, Customers"
          />
        </app-form-field>

        <div class="flex justify-end gap-3">
          <button
            type="button"
            (click)="closeSchemaModal()"
            class="inline-flex items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium text-foreground shadow-sm hover:bg-accent"
          >
            Cancel
          </button>
          <button
            type="button"
            (click)="createSchema()"
            class="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90"
          >
            Create
          </button>
        </div>
      </div>
    </app-modal>

    <app-modal [(isOpen)]="showFieldModal" title="Add field" (backdropClick)="closeFieldModal()">
      <div class="flex flex-col gap-4">
        <app-form-field [label]="'Field key'" [for]="'field-key-input'">
          <app-text-input
            id="field-key-input"
            [value]="modalFieldKey()"
            (valueChange)="modalFieldKey.set($event)"
            placeholder="e.g. title, priority, due_date"
          />
        </app-form-field>

        <app-form-field [label]="'Field type'" [for]="'field-type-select'">
          <app-select-input
            id="field-type-select"
            [options]="fieldTypeOptions()"
            [(value)]="modalFieldType"
          />
        </app-form-field>

        <div class="flex justify-end gap-3">
          <button
            type="button"
            (click)="closeFieldModal()"
            class="inline-flex items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium text-foreground shadow-sm hover:bg-accent"
          >
            Cancel
          </button>
          <button
            type="button"
            (click)="addField()"
            class="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90"
          >
            Add
          </button>
        </div>
      </div>
    </app-modal>
  `,
})
export class ProjectSettingsPageComponent {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly projectService = inject(ProjectService);
  private readonly projectRoute = inject(ProjectRouteService);
  private readonly fieldService = inject(FieldService);

  protected readonly tabs = TABS;
  protected readonly project = this.projectRoute.project;
  protected readonly projectSlug = this.projectRoute.projectSlug;

  protected readonly schemas = this.fieldService.schemas;
  protected readonly activeSchema = this.fieldService.activeSchema;

  private readonly tabParam = toSignal(
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd),
      map(() => this.route.snapshot.paramMap.get('tab')),
      startWith(this.route.snapshot.paramMap.get('tab')),
    ),
    { initialValue: this.route.snapshot.paramMap.get('tab') },
  );

  private readonly subTabParam = toSignal(
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd),
      map(() => this.route.snapshot.paramMap.get('subTab')),
      startWith(this.route.snapshot.paramMap.get('subTab')),
    ),
    { initialValue: this.route.snapshot.paramMap.get('subTab') },
  );

  protected readonly activeTab = computed(() => {
    const tab = this.tabParam();
    if (!tab) return 'general';
    return TABS.some(t => t.id === tab) ? tab : 'general';
  });

  protected readonly activeSchemaKey = computed(() => {
    const tab = this.activeTab();
    const sub = this.subTabParam();
    if (tab === 'datafields' && sub) {
      return sub;
    }
    return null;
  });

  protected readonly nameInputId = 'project-name-input';
  protected readonly projectName = signal('');

  protected readonly showSchemaModal = signal(false);
  protected readonly modalSchemaKey = signal('');
  protected readonly modalSchemaName = signal('');

  protected readonly showFieldModal = signal(false);
  protected readonly modalFieldKey = signal('');
  protected readonly modalFieldType = model<FieldType>('string');

  protected readonly fieldTypeOptions = computed(() =>
    FIELD_TYPES.map(t => ({ value: t, label: t })),
  );

  constructor() {
    effect(
      () => {
        const current = this.project();
        if (current) this.projectName.set(current.name);
      },
      { allowSignalWrites: true },
    );

    effect(
      () => {
        const slug = this.projectSlug();
        const tab = this.activeTab();
        if (slug && tab === 'datafields') {
          this.fieldService.loadSchemas(slug);
        }
      },
      { allowSignalWrites: true },
    );

    effect(
      () => {
        const slug = this.projectSlug();
        const schemaKey = this.activeSchemaKey();
        if (slug && schemaKey) {
          this.fieldService.loadSchema(slug, schemaKey);
        }
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

  protected openSchemaModal(): void {
    this.modalSchemaKey.set('');
    this.modalSchemaName.set('');
    this.showSchemaModal.set(true);
  }

  protected closeSchemaModal(): void {
    this.showSchemaModal.set(false);
  }

  protected createSchema(): void {
    const slug = this.projectSlug();
    const key = this.modalSchemaKey().trim();
    const name = this.modalSchemaName().trim();

    if (!slug || !key || !name) return;

    this.fieldService.createSchema({ project_slug: slug, key, name }).subscribe({
      next: () => {
        this.closeSchemaModal();
        this.fieldService.loadSchemas(slug);
      },
      error: err => console.error('Failed to create schema:', err),
    });
  }

  protected deleteSchema(schemaKey: string): void {
    const slug = this.projectSlug();
    if (!slug) return;

    if (!window.confirm(`Are you sure you want to delete the "${schemaKey}" schema? This action cannot be undone.`)) {
      return;
    }

    this.fieldService.deleteSchema(slug, schemaKey).subscribe({
      next: () => this.fieldService.loadSchemas(slug),
      error: err => console.error('Failed to delete schema:', err),
    });
  }

  protected goToSchema(schemaKey: string): void {
    const slug = this.projectSlug();
    if (!slug) return;
    void this.router.navigate(['/', slug, 'settings', 'datafields', schemaKey]);
  }

  protected goToSchemaList(): void {
    const slug = this.projectSlug();
    if (!slug) return;
    void this.router.navigate(['/', slug, 'settings', 'datafields']);
  }

  protected openFieldModal(): void {
    this.modalFieldKey.set('');
    this.modalFieldType.set('string');
    this.showFieldModal.set(true);
  }

  protected closeFieldModal(): void {
    this.showFieldModal.set(false);
  }

  protected addField(): void {
    const slug = this.projectSlug();
    const schemaKey = this.activeSchemaKey();
    const key = this.modalFieldKey().trim();
    const type = this.modalFieldType();

    if (!slug || !schemaKey || !key) return;

    this.fieldService.createField(slug, schemaKey, { key, type }).subscribe({
      next: () => {
        this.closeFieldModal();
        this.fieldService.loadSchema(slug, schemaKey);
      },
      error: err => console.error('Failed to create field:', err),
    });
  }

  protected removeField(fieldKey: string): void {
    const slug = this.projectSlug();
    const schemaKey = this.activeSchemaKey();
    if (!slug || !schemaKey) return;

    this.fieldService.deleteField(slug, schemaKey, fieldKey).subscribe({
      next: () => this.fieldService.loadSchema(slug, schemaKey),
      error: err => console.error('Failed to delete field:', err),
    });
  }
}

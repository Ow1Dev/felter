import { Component, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { LucideAngularModule } from 'lucide-angular';
import { FormFieldComponent } from '../ui/form-field/form-field';
import { TextInputComponent } from '../ui/text-input/text-input';

@Component({
  selector: 'app-create-project-drawer',
  standalone: true,
  imports: [FormsModule, LucideAngularModule, FormFieldComponent, TextInputComponent],
  host: {
    class: 'block',
  },
  styles: [`
    .drawer-overlay {
      transition: opacity 0.3s ease;
    }
    .drawer-panel {
      transition: transform 0.3s ease;
    }
  `],
  template: `
    <!-- Overlay -->
    <div
      class="drawer-overlay fixed inset-0 bg-black/50 z-40"
      [class.opacity-0]="!isOpen()"
      [class.pointer-events-none]="!isOpen()"
      (click)="close()"
    ></div>

    <!-- Drawer panel -->
    <div
      class="drawer-panel fixed top-0 right-0 h-full w-full max-w-md border-l border-sidebar-border z-50 flex flex-col"
      [class.translate-x-full]="!isOpen()"
      [class.translate-x-0]="isOpen()"
      style="background-color: var(--background);"
    >
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-sidebar-border px-6 py-4">
        <h2 class="text-lg font-semibold text-foreground">New Project</h2>
        <button
          type="button"
          class="rounded-md p-2 text-muted-foreground hover:text-foreground"
          (click)="close()"
        >
          <lucide-icon name="x" [size]="20" />
        </button>
      </div>

      <!-- Form -->
      <div class="flex-1 overflow-auto p-6 flex flex-col gap-6">
        <app-form-field label="Name" [for]="'project-name'" [required]="true">
          <app-text-input
            id="project-name"
            placeholder="Enter project name..."
            [(value)]="name"
            [autocomplete]="'off'"
          />
        </app-form-field>

        <app-form-field label="Description" [for]="'project-desc'">
          <app-text-input
            id="project-desc"
            placeholder="Enter project description..."
            [(value)]="description"
            [autocomplete]="'off'"
          />
        </app-form-field>
      </div>

      <!-- Footer -->
      <div class="border-t border-sidebar-border px-6 py-4 flex justify-end gap-3">
        <button
          type="button"
          class="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-accent"
          (click)="close()"
        >
          Cancel
        </button>
        <button
          type="button"
          class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          [disabled]="!name().trim()"
          (click)="submit()"
        >
          Create Project
        </button>
      </div>
    </div>
  `,
})
export class CreateProjectDrawerComponent {
  readonly isOpen = signal(false);
  readonly name = signal('');
  readonly description = signal('');

  readonly create = output<{ name: string; description: string }>();
  readonly closeDrawer = output<void>();

  open(): void {
    this.name.set('');
    this.description.set('');
    this.isOpen.set(true);
  }

  close(): void {
    this.isOpen.set(false);
    this.closeDrawer.emit();
  }

  submit(): void {
    const trimmedName = this.name().trim();
    if (!trimmedName) return;
    this.create.emit({
      name: trimmedName,
      description: this.description().trim() || '',
    });
    this.isOpen.set(false);
  }
}

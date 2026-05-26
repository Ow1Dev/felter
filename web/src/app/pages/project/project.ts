import { Component, input } from '@angular/core';
import { LucideAngularModule } from 'lucide-angular';

/** Placeholder page for a specific project. */
@Component({
  selector: 'app-project-page',
  standalone: true,
  imports: [LucideAngularModule],
  host: {
    class: 'flex flex-col w-full h-full',
  },
  template: `
    <div class="flex flex-1 flex-col items-center justify-center gap-4 text-center p-8">
      <lucide-icon name="hard-hat" [size]="48" class="text-muted-foreground/40" />
      <div class="flex flex-col gap-2">
        <h1 class="text-2xl font-bold text-foreground">Project: {{ slug() }}</h1>
        <p class="text-sm text-muted-foreground max-w-md">
          This is a placeholder for the project detail page. Replace with real implementation.
        </p>
      </div>
    </div>
  `,
})
export class ProjectPageComponent {
  readonly slug = input.required<string>();
}

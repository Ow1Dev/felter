import { Component, computed, HostListener, input, model, signal } from '@angular/core';

export interface SelectOption {
  value: string;
  label: string;
}

/** Reusable select input with styling aligned to the design system.
 *  Renders a custom dropdown so options are fully styleable in dark mode. */
@Component({
  selector: 'app-select-input',
  standalone: true,
  host: {
    class: 'block w-full',
  },
  template: `
    <div class="relative" #container>
      <button
        type="button"
        (click)="toggle()"
        class="w-full rounded-md border px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring flex items-center justify-between"
        style="background-color: var(--popover); border-color: var(--border); color: var(--popover-foreground);"
        [class.ring-2]="isOpen()"
        [class.ring-ring]="isOpen()"
      >
        <span>{{ selectedLabel() }}</span>
        <svg
          class="h-4 w-4 opacity-50 transition-transform"
          [class.rotate-180]="isOpen()"
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      @if (isOpen()) {
        <div
          class="absolute z-50 mt-1 w-full rounded-md border shadow-lg overflow-hidden"
          style="background-color: var(--popover); border-color: var(--border);"
        >
          @for (opt of options(); track opt.value) {
            <button
              type="button"
              (click)="select(opt.value)"
              class="w-full px-3 py-2 text-sm text-left transition-colors"
              style="color: var(--popover-foreground);"
              [class.hover:bg-accent]="value() !== opt.value"
              [style.background-color]="value() === opt.value ? 'var(--accent)' : ''"
            >
              {{ opt.label }}
            </button>
          }
        </div>
      }
    </div>
  `,
})
export class SelectInputComponent {
  readonly id = input<string | null>(null);
  readonly disabled = input<boolean>(false);
  readonly options = input<SelectOption[]>([]);

  readonly value = model<string>('');

  protected readonly isOpen = signal(false);

  protected readonly selectedLabel = computed(() => {
    const opt = this.options().find(o => o.value === this.value());
    return opt?.label ?? '';
  });

  protected toggle(): void {
    if (this.disabled()) return;
    this.isOpen.update(v => !v);
  }

  protected select(val: string): void {
    this.value.set(val);
    this.isOpen.set(false);
  }

  @HostListener('document:click', ['$event'])
  protected onDocumentClick(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    if (!target.closest('app-select-input')) {
      this.isOpen.set(false);
    }
  }
}

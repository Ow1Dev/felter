import { Component, computed, input, model, output } from '@angular/core';

/** Reusable text input with styling aligned to the design system. */
@Component({
  selector: 'app-text-input',
  standalone: true,
  host: {
    class: 'block w-full',
  },
  template: `
    <input
      [id]="id() ?? undefined"
      [attr.type]="type()"
      [attr.placeholder]="placeholder() ?? undefined"
      [attr.autocomplete]="autocomplete() ?? undefined"
      [disabled]="disabled()"
      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
      [value]="value()"
      (input)="onInput($event)"
      (blur)="onBlur()"
    />
  `,
})
export class TextInputComponent {
  readonly id = input<string | null>(null);
  readonly placeholder = input<string | null>(null);
  readonly autocomplete = input<string | null>(null);
  readonly disabled = input<boolean>(false);
  readonly type = input<string>('text');

  readonly value = model<string>('');

  readonly blur = output<void>();
  readonly valueChange = output<string>();

  protected onInput(event: Event): void {
    const value = (event.target as HTMLInputElement).value;
    this.value.set(value);
    this.valueChange.emit(value);
  }

  protected onBlur(): void {
    this.blur.emit();
  }
}

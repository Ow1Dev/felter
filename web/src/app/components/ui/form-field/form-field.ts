import { NgIf } from '@angular/common';
import { Component, InputSignal, input } from '@angular/core';

/** Shared form field wrapper providing label + description slots. */
@Component({
  selector: 'app-form-field',
  standalone: true,
  imports: [NgIf],
  host: {
    class: 'flex flex-col gap-2',
  },
  template: `
    <div class="flex flex-col gap-1.5">
      <label
        class="text-sm font-medium text-foreground"
        [attr.for]="forId() ?? undefined"
      >
        {{ label() }}
        <span *ngIf="required()" class="text-destructive">*</span>
      </label>
      <ng-content />
      <p *ngIf="description()" class="text-xs text-muted-foreground">
        {{ description() }}
      </p>
    </div>
  `,
})
export class FormFieldComponent {
  readonly label: InputSignal<string> = input.required<string>();
  readonly description = input<string | null>(null);
  readonly forId = input<string | null>(null, { alias: 'for' });
  readonly required = input<boolean>(false);
}

import { Component, input, model, output } from '@angular/core';

/** Reusable centered modal with backdrop.
 *  Uses CSS variables so colors always match the active theme. */
@Component({
  selector: 'app-modal',
  standalone: true,
  host: {
    class: 'block',
  },
  template: `
    @if (isOpen()) {
      <div
        class="fixed inset-0 z-50 flex items-center justify-center"
        style="background-color: rgba(0, 0, 0, 0.25);"
        (click)="backdropClick.emit()"
      >
        <div
          class="w-full max-w-md rounded-xl p-6 shadow-lg border"
          style="background-color: var(--popover); border-color: var(--border); color: var(--popover-foreground);"
          (click)="$event.stopPropagation()"
        >
          @if (title()) {
            <h3 class="text-lg font-semibold mb-4">{{ title() }}</h3>
          }
          <ng-content />
        </div>
      </div>
    }
  `,
})
export class ModalComponent {
  readonly isOpen = model.required<boolean>();
  readonly title = input<string>('');
  readonly backdropClick = output<void>();
}

import { Component, computed, effect, inject } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { toSignal } from '@angular/core/rxjs-interop';
import { map } from 'rxjs/operators';
import { ViewService } from '../../services/view.service';

/** Temporary placeholder for workspace views until real feature pages exist. */
@Component({
  standalone: true,
  template: `
    <section class="flex flex-1 flex-col gap-6 overflow-auto p-10">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight text-foreground">
          {{ viewName() ?? 'Unknown view' }}
        </h1>
        <p class="mt-2 text-sm text-muted-foreground">
          This is a placeholder for the <span class="font-medium">{{ viewSlug() }}</span>
          view. Replace with the real implementation when ready.
        </p>
      </div>
    </section>
  `,
})
export class ViewPlaceholderComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly viewService = inject(ViewService);

  private readonly paramMap = toSignal(
    this.route.paramMap.pipe(map(params => params.get('viewSlug') ?? '')),
    { initialValue: this.route.snapshot.paramMap.get('viewSlug') ?? '' },
  );

  protected readonly viewSlug = computed(() => this.paramMap());

  readonly viewName = computed(() => this.viewService.getBySlug(this.viewSlug())?.name);

  constructor() {
    effect(() => {
      const slug = this.viewSlug();
      if (slug) this.viewService.setActiveBySlug(slug);
    });
  }
}

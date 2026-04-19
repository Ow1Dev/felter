import { Component, OnInit, signal } from '@angular/core';
import { HelloService } from './services/hello.service';

@Component({
  selector: 'app-root',
  template: `
    <main style="display:flex;align-items:center;justify-content:center;min-height:100dvh;font-family:system-ui,Segoe UI,Arial,sans-serif;">
      <div style="max-width:680px;padding:2rem;text-align:center;">
        <h1 style="margin:0 0 1rem;">Felter</h1>
        <p style="margin:0 0 1.5rem;color:#444;">Angular ↔ Go API demo</p>
        <h2 style="margin:1rem 0;">{{ greeting() || 'Loading…' }}</h2>
        <button (click)="load()" style="padding:.5rem 1rem;border:1px solid #ccc;border-radius:6px;background:#fff;cursor:pointer;">Refresh Greeting</button>
        <details style="margin-top:1rem;">
          <summary>Response JSON</summary>
          <pre style="text-align:left;background:#f6f8fa;border-radius:8px;padding:1rem;margin-top:.5rem;overflow:auto;">{{ raw() }}</pre>
        </details>
      </div>
    </main>
  `,
})
export class App implements OnInit {
  protected readonly raw = signal('');
  protected readonly greeting = signal('');

  constructor(private hello: HelloService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.hello.getHello().subscribe({
      next: (res) => {
        this.greeting.set(res?.message ?? '');
        this.raw.set(JSON.stringify(res, null, 2));
      },
      error: (err) => {
        this.greeting.set('');
        this.raw.set(JSON.stringify(err?.error ?? { error: 'request failed' }, null, 2));
      },
    });
  }
}

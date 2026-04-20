import { Injectable, linkedSignal, signal } from '@angular/core';

export interface Workspace {
  id: string;
  name: string;
  /** Two-letter initials displayed when collapsed or as the avatar. */
  initials: string;
  /** Optional accent colour class for the initials badge. */
  color: string;
}

const MOCK_WORKSPACES: Workspace[] = [
  { id: 'ws-1', name: 'Acme Field Ops',   initials: 'AF', color: 'bg-blue-500' },
  { id: 'ws-2', name: 'Metro Services',   initials: 'MS', color: 'bg-violet-500' },
  { id: 'ws-3', name: 'Coastal Technics', initials: 'CT', color: 'bg-emerald-500' },
];

/** Manages the list of workspaces and tracks the currently active one. */
@Injectable({ providedIn: 'root' })
export class WorkspaceService {
  readonly workspaces = signal<Workspace[]>(MOCK_WORKSPACES);

  readonly activeWorkspace = linkedSignal<Workspace[], Workspace>({
    source: this.workspaces,
    computation: (workspaces, previous) => {
      if (previous) {
        const still = workspaces.find(w => w.id === previous.value.id);
        if (still) return still;
      }
      return workspaces[0];
    },
  });

  /** Switch the active workspace by id. */
  setActive(id: string): void {
    const ws = this.workspaces().find(w => w.id === id);
    if (ws) this.activeWorkspace.set(ws);
  }
}

import { Injectable, linkedSignal, signal } from '@angular/core';

export interface Workspace {
  id: string;
  slug: string;
  name: string;
  /** Two-letter initials displayed when collapsed or as the avatar. */
  initials: string;
  /** Optional accent colour class for the initials badge. */
  color: string;
}

const MOCK_WORKSPACES: Workspace[] = [
  {
    id: 'ws-1',
    slug: 'acme-field-ops',
    name: 'Acme Field Ops',
    initials: 'AF',
    color: 'bg-blue-500',
  },
  {
    id: 'ws-2',
    slug: 'metro-services',
    name: 'Metro Services',
    initials: 'MS',
    color: 'bg-violet-500',
  },
  {
    id: 'ws-3',
    slug: 'coastal-technics',
    name: 'Coastal Technics',
    initials: 'CT',
    color: 'bg-emerald-500',
  },
];

const ACTIVE_WORKSPACE_STORAGE_KEY = 'felter-active-workspace';

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

  constructor() {
    const storedSlug = this.getStoredActiveWorkspaceSlug();
    if (storedSlug) this.setActiveBySlug(storedSlug);
  }

  /** Switch the active workspace by id. */
  setActive(id: string): void {
    const ws = this.workspaces().find(w => w.id === id);
    if (ws) {
      this.activeWorkspace.set(ws);
      this.persistActiveWorkspace(ws.slug);
    }
  }

  /** Switch the active workspace by slug. */
  setActiveBySlug(slug: string): void {
    const ws = this.getBySlug(slug);
    if (ws) {
      this.activeWorkspace.set(ws);
      this.persistActiveWorkspace(ws.slug);
    }
  }

  /** Update a workspace's name by slug. */
  updateWorkspaceName(slug: string, name: string): void {
    this.workspaces.update(list =>
      list.map(workspace =>
        workspace.slug === slug ? { ...workspace, name: name.trim() } : workspace,
      ),
    );
  }

  /** Get the default workspace. */
  getDefaultWorkspace(): Workspace | null {
    return this.workspaces()[0] ?? null;
  }

  /** Get a workspace by slug. */
  getBySlug(slug: string): Workspace | undefined {
    return this.workspaces().find(workspace => workspace.slug === slug);
  }

  /** Retrieve the stored active workspace slug, if valid. */
  getStoredActiveWorkspaceSlug(): string | null {
    if (typeof window === 'undefined' || typeof window.localStorage === 'undefined') {
      return null;
    }

    try {
      const slug = window.localStorage.getItem(ACTIVE_WORKSPACE_STORAGE_KEY);
      if (!slug) return null;
      const exists = !!this.getBySlug(slug);
      if (!exists) {
        window.localStorage.removeItem(ACTIVE_WORKSPACE_STORAGE_KEY);
        return null;
      }
      return slug;
    } catch {
      return null;
    }
  }

  private persistActiveWorkspace(slug: string): void {
    if (typeof window === 'undefined' || typeof window.localStorage === 'undefined') {
      return;
    }

    try {
      window.localStorage.setItem(ACTIVE_WORKSPACE_STORAGE_KEY, slug);
    } catch {
      /* ignore storage errors */
    }
  }
}

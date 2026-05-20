import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from '../../environments/environment';
import type { components } from '../api/projectservice';

export type Project = components['schemas']['Project'];
export type CreateProjectRequest = components['schemas']['CreateProjectRequest'];

const ACTIVE_PROJECT_STORAGE_KEY = 'felter-active-project';

function getInitials(name: string): string {
  const parts = name.split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.slice(0, 2).toUpperCase();
}

const COLOR_CLASSES = [
  'bg-blue-500',
  'bg-violet-500',
  'bg-emerald-500',
  'bg-orange-500',
  'bg-pink-500',
  'bg-cyan-500',
  'bg-red-500',
  'bg-indigo-500',
  'bg-amber-500',
  'bg-teal-500',
];

function getColorClass(slug: string): string {
  let hash = 0;
  for (let i = 0; i < slug.length; i++) {
    hash = slug.charCodeAt(i) + ((hash << 5) - hash);
  }
  return COLOR_CLASSES[Math.abs(hash) % COLOR_CLASSES.length];
}

/** UI representation of a project with display helpers. */
export interface ProjectDisplay {
  project: Project;
  initials: string;
  color: string;
}

@Injectable({ providedIn: 'root' })
export class ProjectService {
  readonly projects = signal<Project[]>([]);
  readonly activeProject = signal<Project | null>(null);

  constructor(private http: HttpClient) {}

  listProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${environment.projectsUrl}/projects`);
  }

  createProject(request: CreateProjectRequest): Observable<Project> {
    return this.http.post<Project>(`${environment.projectsUrl}/projects`, request);
  }

  loadProjects(): void {
    this.listProjects().subscribe({
      next: projects => this.projects.set(projects),
      error: err => console.error('Failed to load projects:', err),
    });
  }

  loadProjectsAndSetActive(slug: string): void {
    this.listProjects().subscribe({
      next: projects => {
        this.projects.set(projects);
        const found = projects.find(p => p.slug === slug);
        if (found) {
          this.activeProject.set(found);
        }
      },
      error: err => console.error('Failed to load projects:', err),
    });
  }

  setActive(project: Project): void {
    this.activeProject.set(project);
    this.persistActiveProject(project.slug);
  }

  setActiveBySlug(slug: string): void {
    const found = this.projects().find(p => p.slug === slug);
    if (found) {
      this.activeProject.set(found);
      this.persistActiveProject(slug);
    }
  }

  getBySlug(slug: string): Project | undefined {
    return this.projects().find(p => p.slug === slug);
  }

  searchProjects(query: string): Project[] {
    if (!query.trim()) return this.projects();
    const lower = query.toLowerCase();
    return this.projects().filter(p =>
      p.name.toLowerCase().includes(lower) ||
      p.slug.toLowerCase().includes(lower),
    );
  }

  toDisplay(project: Project): ProjectDisplay {
    return {
      project,
      initials: getInitials(project.name),
      color: getColorClass(project.slug),
    };
  }

  getDefaultProject(): Project | null {
    return this.projects()[0] ?? null;
  }

  getStoredActiveProjectSlug(): string | null {
    if (typeof window === 'undefined' || typeof window.localStorage === 'undefined') {
      return null;
    }
    try {
      const slug = window.localStorage.getItem(ACTIVE_PROJECT_STORAGE_KEY);
      if (!slug) return null;
      return slug;
    } catch {
      return null;
    }
  }

  private persistActiveProject(slug: string): void {
    if (typeof window === 'undefined' || typeof window.localStorage === 'undefined') {
      return;
    }
    try {
      window.localStorage.setItem(ACTIVE_PROJECT_STORAGE_KEY, slug);
    } catch {
      /* ignore storage errors */
    }
  }
}

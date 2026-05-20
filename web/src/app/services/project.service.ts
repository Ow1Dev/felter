import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import type { components } from '../api/projectservice';

export type Project = components['schemas']['Project'];
export type CreateProjectRequest = components['schemas']['CreateProjectRequest'];

@Injectable({ providedIn: 'root' })
export class ProjectService {
  constructor(private http: HttpClient) {}

  listProjects(): Observable<Project[]> {
    return this.http.get<Project[]>(`${environment.projectsUrl}/projects`);
  }

  createProject(request: CreateProjectRequest): Observable<Project> {
    return this.http.post<Project>(`${environment.projectsUrl}/projects`, request);
  }
}

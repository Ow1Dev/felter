import { Injectable, signal } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import type { components } from '../api/fieldservice';

export type Schema = components['schemas']['Schema'];
export type SchemaWithFields = components['schemas']['SchemaWithFields'];
export type SchemaField = components['schemas']['SchemaField'];
export type CreateSchemaRequest = components['schemas']['CreateSchemaRequest'];
export type CreateSchemaFieldRequest = components['schemas']['CreateSchemaFieldRequest'];
export type FieldType = components['schemas']['FieldType'];

const FIELD_TYPES: FieldType[] = ['string', 'int', 'float', 'boolean', 'date', 'datetime'];

export { FIELD_TYPES };

@Injectable({ providedIn: 'root' })
export class FieldService {
  readonly schemas = signal<Schema[]>([]);
  readonly activeSchema = signal<SchemaWithFields | null>(null);

  constructor(private http: HttpClient) {}

  listSchemas(projectSlug: string): Observable<Schema[]> {
    const params = new HttpParams().set('project_slug', projectSlug);
    return this.http.get<Schema[]>(`${environment.fieldUrl}/schemas`, { params });
  }

  loadSchemas(projectSlug: string): void {
    this.listSchemas(projectSlug).subscribe({
      next: schemas => this.schemas.set(schemas),
      error: err => console.error('Failed to load schemas:', err),
    });
  }

  createSchema(request: CreateSchemaRequest): Observable<Schema> {
    return this.http.post<Schema>(`${environment.fieldUrl}/schemas`, request);
  }

  getSchema(projectSlug: string, schemaKey: string): Observable<SchemaWithFields> {
    const params = new HttpParams().set('project_slug', projectSlug);
    return this.http.get<SchemaWithFields>(`${environment.fieldUrl}/schemas/${schemaKey}`, { params });
  }

  loadSchema(projectSlug: string, schemaKey: string): void {
    this.activeSchema.set(null);
    this.getSchema(projectSlug, schemaKey).subscribe({
      next: schema => this.activeSchema.set(schema),
      error: err => console.error('Failed to load schema:', err),
    });
  }

  deleteSchema(projectSlug: string, schemaKey: string): Observable<void> {
    const params = new HttpParams().set('project_slug', projectSlug);
    return this.http.delete<void>(`${environment.fieldUrl}/schemas/${schemaKey}`, { params });
  }

  createField(projectSlug: string, schemaKey: string, request: CreateSchemaFieldRequest): Observable<SchemaField> {
    const params = new HttpParams().set('project_slug', projectSlug);
    return this.http.post<SchemaField>(`${environment.fieldUrl}/schemas/${schemaKey}/fields`, request, { params });
  }

  deleteField(projectSlug: string, schemaKey: string, fieldKey: string): Observable<void> {
    const params = new HttpParams().set('project_slug', projectSlug);
    return this.http.delete<void>(`${environment.fieldUrl}/schemas/${schemaKey}/fields/${fieldKey}`, { params });
  }
}

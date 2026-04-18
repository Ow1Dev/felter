import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface HelloResponse {
  message: string;
}

@Injectable({ providedIn: 'root' })
export class HelloService {
  constructor(private http: HttpClient) {}

  getHello(): Observable<HelloResponse> {
    return this.http.get<HelloResponse>(`${environment.apiBaseUrl}/api/hello`);
  }
}

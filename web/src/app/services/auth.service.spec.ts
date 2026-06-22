import { TestBed } from '@angular/core/testing';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { AuthService, CurrentUser } from './auth.service';
import { authInterceptor } from '../auth.interceptor';

describe('AuthService', () => {
  const TOKEN_KEY = 'felter_auth_token';
  let service: AuthService;
  let httpMock: HttpTestingController;

  const mockUser: CurrentUser = {
    id: 1,
    email: 'test@example.com',
    username: 'test',
    display_name: 'Test User',
    created_at: '2024-01-01T00:00:00Z',
  };

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
      ],
    });
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  describe('with stored token', () => {
    beforeEach(() => {
      localStorage.setItem(TOKEN_KEY, 'stored-token');
      service = TestBed.inject(AuthService);
    });

    it('should call /me via queueMicrotask and set currentUser on success', async () => {
      // Wait for the queueMicrotask deferral to flush.
      await Promise.resolve();

      const req = httpMock.expectOne(req => req.url.endsWith('/me'));
      expect(req.request.headers.get('Authorization')).toBe('Bearer stored-token');
      req.flush(mockUser);

      await service.waitForInit();

      expect(service.currentUser()).toEqual(mockUser);
      expect(service.isAuthenticated()).toBe(true);
      expect(localStorage.getItem(TOKEN_KEY)).toBe('stored-token');
    });

    it('should clear token on 401 from /me', async () => {
      await Promise.resolve();

      const req = httpMock.expectOne(req => req.url.endsWith('/me'));
      req.flush({ error: 'unauthorized' }, { status: 401, statusText: 'Unauthorized' });

      await service.waitForInit();

      expect(service.currentUser()).toBeNull();
      expect(service.token()).toBeNull();
      expect(localStorage.getItem(TOKEN_KEY)).toBeNull();
    });

    it('should NOT clear token on non-401 errors from /me', async () => {
      await Promise.resolve();

      const req = httpMock.expectOne(req => req.url.endsWith('/me'));
      req.flush({ error: 'server error' }, { status: 500, statusText: 'Internal Server Error' });

      await service.waitForInit();

      expect(service.currentUser()).toBeNull();
      expect(service.token()).toBe('stored-token');
      expect(localStorage.getItem(TOKEN_KEY)).toBe('stored-token');
    });
  });

  describe('without stored token', () => {
    beforeEach(() => {
      service = TestBed.inject(AuthService);
    });

    it('should resolve waitForInit immediately and not call /me', async () => {
      await service.waitForInit();

      httpMock.expectNone(req => req.url.endsWith('/me'));
      expect(service.isAuthenticated()).toBe(false);
      expect(service.currentUser()).toBeNull();
    });
  });

  describe('handleCallback', () => {
    beforeEach(() => {
      localStorage.setItem(TOKEN_KEY, 'old-token');
      service = TestBed.inject(AuthService);
    });

    it('should exchange code for token and fetch user', async () => {
      // Drain the restoreSession microtask first.
      await Promise.resolve();
      httpMock.expectOne(req => req.url.endsWith('/me')).flush(mockUser);
      await service.waitForInit();

      // Simulate callback URL with code
      const urlSpy = vi.spyOn(window, 'location', 'get').mockReturnValue({
        ...window.location,
        search: '?code=auth-code',
      } as Location);

      const promise = new Promise<void>((resolve, reject) => {
        service.handleCallback().subscribe({
          next: resp => {
            expect(resp.token).toBe('new-token');
            expect(resp.email).toBe('test@example.com');
            resolve();
          },
          error: reject,
        });
      });

      const req = httpMock.expectOne(req => req.url.endsWith('/callback'));
      expect(req.request.body).toEqual({ code: 'auth-code' });
      req.flush({ token: 'new-token', email: 'test@example.com' });

      // The callback handler chains fetchCurrentUser
      const meReq = httpMock.expectOne(req => req.url.endsWith('/me'));
      meReq.flush(mockUser);

      await promise;

      expect(service.token()).toBe('new-token');
      expect(localStorage.getItem(TOKEN_KEY)).toBe('new-token');

      urlSpy.mockRestore();
    });
  });
});

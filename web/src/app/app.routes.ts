import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () =>
      import('./pages/login/login-page').then(m => m.LoginPageComponent),
  },
  {
    path: 'callback',
    loadComponent: () =>
      import('./pages/callback/callback-page').then(m => m.CallbackPageComponent),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/workspace-selector/workspace-selector').then(m => m.WorkspaceSelectorComponent),
  },
  {
    path: 'no-workspace',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/no-workspace/no-workspace').then(m => m.NoWorkspacePageComponent),
  },
  {
    path: 'settings',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/user-settings/user-settings-page').then(
        m => m.UserSettingsPageComponent,
      ),
  },
  {
    path: ':workspaceSlug',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./layouts/workspace-layout/workspace-layout').then(
        m => m.WorkspaceLayoutComponent,
      ),
    children: [
      {
        path: '',
        pathMatch: 'full',
        loadComponent: () =>
          import('./pages/workspace-redirect/workspace-redirect').then(
            m => m.WorkspaceRedirectComponent,
          ),
      },
      {
        path: 'views/:viewSlug',
        loadComponent: () =>
          import('./pages/view-placeholder/view-placeholder').then(
            m => m.ViewPlaceholderComponent,
          ),
      },
      {
        path: 'settings',
        children: [
          {
            path: '',
            pathMatch: 'full',
            loadComponent: () =>
              import('./pages/workspace-settings/workspace-settings-page').then(
                m => m.WorkspaceSettingsPageComponent,
              ),
          },
          {
            path: ':tab',
            loadComponent: () =>
              import('./pages/workspace-settings/workspace-settings-page').then(
                m => m.WorkspaceSettingsPageComponent,
              ),
          },
        ],
      },
    ],
  },
  {
    path: '**',
    redirectTo: '',
  },
];

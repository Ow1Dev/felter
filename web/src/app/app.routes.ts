import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./pages/root-redirect/root-redirect').then(m => m.RootRedirectComponent),
  },
  {
    path: 'no-workspace',
    loadComponent: () =>
      import('./pages/no-workspace/no-workspace').then(m => m.NoWorkspacePageComponent),
  },
  {
    path: 'settings',
    loadComponent: () =>
      import('./pages/user-settings/user-settings-page').then(
        m => m.UserSettingsPageComponent,
      ),
  },
  {
    path: ':workspaceSlug',
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

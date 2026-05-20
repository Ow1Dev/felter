import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  {
    path: 'callback',
    loadComponent: () =>
      import('./pages/callback/callback-page').then(m => m.CallbackPageComponent),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/project-selector/project-selector').then(m => m.ProjectSelectorComponent),
  },
  {
    path: 'no-project',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/no-project/no-project').then(m => m.NoProjectPageComponent),
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
    path: 'projects/:slug',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/project/project').then(m => m.ProjectPageComponent),
  },
  {
    path: ':projectSlug',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./layouts/project-layout/project-layout').then(
        m => m.ProjectLayoutComponent,
      ),
    children: [
      {
        path: '',
        pathMatch: 'full',
        loadComponent: () =>
          import('./pages/project-redirect/project-redirect').then(
            m => m.ProjectRedirectComponent,
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
              import('./pages/project-settings/project-settings-page').then(
                m => m.ProjectSettingsPageComponent,
              ),
          },
          {
            path: ':tab',
            loadComponent: () =>
              import('./pages/project-settings/project-settings-page').then(
                m => m.ProjectSettingsPageComponent,
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

import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import {
  Briefcase,
  Building2,
  Calendar,
  Check,
  ChevronDown,
  Ellipsis,
  FileText,
  HardHat,
  LUCIDE_ICONS,
  LucideIconProvider,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
  Settings,
  Sun,
  Search,
  UserRound,
  Users,
  Wrench,
  LogOut,
  X,
} from 'lucide-angular';

import { routes } from './app.routes';
import { authInterceptor } from './auth.interceptor';

/** All Lucide icons used in the sidebar. */
const icons = {
  Briefcase,
  Building2,
  Calendar,
  Check,
  ChevronDown,
  Ellipsis,
  FileText,
  HardHat,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
  Settings,
  LogOut,
  Search,
  Sun,
  UserRound,
  Users,
  Wrench,
  X,
};

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideHttpClient(withInterceptors([authInterceptor])),
    { provide: LUCIDE_ICONS, multi: true, useValue: new LucideIconProvider(icons) },
  ],
};

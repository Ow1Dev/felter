import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import {
  Briefcase, Building2, Calendar, Check, ChevronDown, Ellipsis,
  FileText, HardHat, LUCIDE_ICONS, LucideIconProvider, Moon,
  PanelLeftClose, PanelLeftOpen, Search, Settings, Sun, UserRound, Users, Wrench,
} from 'lucide-angular';
import { App } from './app';

const testIcons = {
  Briefcase, Building2, Calendar, Check, ChevronDown, Ellipsis,
  FileText, HardHat, Moon, PanelLeftClose, PanelLeftOpen, Search,
  Settings, Sun, UserRound, Users, Wrench,
};

describe('App', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [App],
      providers: [
        provideRouter([]),
        { provide: LUCIDE_ICONS, multi: true, useValue: new LucideIconProvider(testIcons) },
      ],
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(App);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });

  it('should render the app header', () => {
    const fixture = TestBed.createComponent(App);
    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.querySelector('app-app-header')).toBeTruthy();
  });
});

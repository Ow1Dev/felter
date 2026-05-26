import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, Router } from '@angular/router';
import { BehaviorSubject } from 'rxjs';
import { vi } from 'vitest';

import { ProjectSettingsPageComponent } from './project-settings-page';
import { ProjectService, type Project } from '../../services/project.service';
import { ProjectRouteService } from '../../services/project-route.service';

class RouterStub {
  navigate = vi.fn().mockResolvedValue(true);
}

function createActivatedRouteStub(): ActivatedRoute & { setTab(tab: string | null): void } {
  const paramMapSubject = new BehaviorSubject(convertToParamMap({}));
  const snapshot: any = { paramMap: paramMapSubject.value };

  const stub: any = {
    paramMap: paramMapSubject.asObservable(),
    snapshot,
    setTab(tab: string | null) {
      const map = convertToParamMap(tab ? { tab } : {});
      paramMapSubject.next(map);
      snapshot.paramMap = map;
    },
  };

  return stub;
}

function provideProjectRouteServiceStub() {
  return {
    provide: ProjectRouteService,
    useFactory: (projectService: ProjectService) => ({
      project: () => projectService.activeProject(),
      projectSlug: () => projectService.activeProject()?.slug ?? null,
    }),
    deps: [ProjectService],
  } as const;
}

const mockProject: Project = {
  id: '1',
  slug: 'test-project',
  name: 'Test Project',
  description: 'A test project',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe('ProjectSettingsPageComponent', () => {
  let projectService: ProjectService;
  let routeStub: ActivatedRoute & { setTab(tab: string | null): void };

  beforeEach(async () => {
    routeStub = createActivatedRouteStub();
    routeStub.setTab(null);

    await TestBed.configureTestingModule({
      imports: [ProjectSettingsPageComponent],
      providers: [
        ProjectService,
        { provide: Router, useClass: RouterStub },
        provideProjectRouteServiceStub(),
        { provide: ActivatedRoute, useValue: routeStub },
      ],
    }).compileComponents();

    projectService = TestBed.inject(ProjectService);
    projectService.projects.set([mockProject]);
    projectService.activeProject.set(mockProject);
  });

  it('renders the general tab by default', () => {
    const fixture = TestBed.createComponent(ProjectSettingsPageComponent);
    fixture.detectChanges();

    const heading = fixture.nativeElement.querySelector('[data-testid="project-title-heading"]');
    expect(heading?.textContent).toContain('Project title');
  });

  it('shows data tab content when navigating to /datafields', () => {
    const fixture = TestBed.createComponent(ProjectSettingsPageComponent);
    fixture.detectChanges();

    routeStub.setTab('datafields');
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Data settings');
  });
});

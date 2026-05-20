import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, Router } from '@angular/router';
import { BehaviorSubject } from 'rxjs';
import { vi } from 'vitest';

import { WorkspaceSettingsPageComponent } from './workspace-settings-page';
import { WorkspaceService } from '../../services/workspace.service';
import { WorkspaceRouteService } from '../../services/workspace-route.service';

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

function provideWorkspaceRouteServiceStub() {
  return {
    provide: WorkspaceRouteService,
    useFactory: (workspaceService: WorkspaceService) => ({
      workspace: () => workspaceService.activeWorkspace(),
      workspaceSlug: () => workspaceService.activeWorkspace().slug,
    }),
    deps: [WorkspaceService],
  } as const;
}

describe('WorkspaceSettingsPageComponent', () => {
  let workspaceService: WorkspaceService;
  let routeStub: ActivatedRoute & { setTab(tab: string | null): void };

  beforeEach(async () => {
    routeStub = createActivatedRouteStub();
    routeStub.setTab(null);

    await TestBed.configureTestingModule({
      imports: [WorkspaceSettingsPageComponent],
      providers: [
        WorkspaceService,
        { provide: Router, useClass: RouterStub },
        provideWorkspaceRouteServiceStub(),
        { provide: ActivatedRoute, useValue: routeStub },
      ],
    }).compileComponents();

    workspaceService = TestBed.inject(WorkspaceService);
  });

  it('renders the general tab by default', () => {
    const fixture = TestBed.createComponent(WorkspaceSettingsPageComponent);
    fixture.detectChanges();

    const heading = fixture.nativeElement.querySelector('[data-testid="workspace-title-heading"]');
    expect(heading?.textContent).toContain('Workspace title');
  });

  it('updates the workspace name when the field blurs', () => {
    const fixture = TestBed.createComponent(WorkspaceSettingsPageComponent);
    fixture.detectChanges();

    const input: HTMLInputElement = fixture.nativeElement.querySelector('input');
    input.value = 'Renamed Workspace';
    input.dispatchEvent(new Event('input'));
    input.dispatchEvent(new Event('blur'));

    expect(workspaceService.activeWorkspace().name).toBe('Renamed Workspace');
  });

  it('shows data tab content when navigating to /datafields', () => {
    const fixture = TestBed.createComponent(WorkspaceSettingsPageComponent);
    fixture.detectChanges();

    routeStub.setTab('datafields');
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Data settings');
  });
});

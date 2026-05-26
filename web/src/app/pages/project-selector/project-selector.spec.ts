import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LUCIDE_ICONS, LucideIconProvider, Plus, Search, X } from 'lucide-angular';
import { ProjectSelectorComponent } from './project-selector';
import { ProjectService } from '../../services/project.service';
import { ViewService } from '../../services/view.service';
import { Router } from '@angular/router';
import { signal } from '@angular/core';
import { of } from 'rxjs';
import { vi } from 'vitest';

describe('ProjectSelectorComponent', () => {
  let component: ProjectSelectorComponent;
  let fixture: ComponentFixture<ProjectSelectorComponent>;
  let mockProjectService: Partial<ProjectService>;
  let mockViewService: Partial<ViewService>;
  let mockRouter: Partial<Router>;

  beforeEach(async () => {
    mockProjectService = {
      loadProjects: vi.fn(),
      searchProjects: vi.fn().mockReturnValue([]),
      toDisplay: vi.fn(),
      getBySlug: vi.fn(),
      setActive: vi.fn(),
      createProject: vi.fn(),
      projects: signal([]),
      activeProject: signal(null),
    };

    mockViewService = {
      getDefaultView: vi.fn(),
    };

    mockRouter = {
      navigate: vi.fn().mockResolvedValue(true),
    };

    await TestBed.configureTestingModule({
      imports: [ProjectSelectorComponent],
      providers: [
        { provide: ProjectService, useValue: mockProjectService },
        { provide: ViewService, useValue: mockViewService },
        { provide: Router, useValue: mockRouter },
        { provide: LUCIDE_ICONS, multi: true, useValue: new LucideIconProvider({ Plus, Search, X }) },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ProjectSelectorComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should load projects on init', () => {
    fixture.detectChanges();
    expect(mockProjectService.loadProjects).toHaveBeenCalled();
  });

  it('should render the project grid when projects exist', () => {
    const mockProject = {
      id: '1',
      name: 'Test Project',
      slug: 'test-project',
      description: 'A test project',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    (mockProjectService.searchProjects as ReturnType<typeof vi.fn>).mockReturnValue([mockProject]);
    (mockProjectService.toDisplay as ReturnType<typeof vi.fn>).mockReturnValue({
      project: mockProject,
      initials: 'TP',
      color: 'bg-blue-500',
    });

    fixture.detectChanges();

    const grid = fixture.nativeElement.querySelector('.grid');
    expect(grid).toBeTruthy();

    const projectButtons = fixture.nativeElement.querySelectorAll('button');
    expect(projectButtons.length).toBeGreaterThan(0);
  });

  it('should show empty state when no projects exist', () => {
    fixture.detectChanges();

    const emptyStateText = Array.from(fixture.nativeElement.querySelectorAll('p')).find((p: any) =>
      p.textContent?.includes('We couldn\'t find any projects'),
    );
    expect(emptyStateText).toBeTruthy();
  });

  it('should not have a refresh button in the empty state', () => {
    fixture.detectChanges();

    const emptyStateContainer = Array.from(fixture.nativeElement.querySelectorAll('div')).find((div: any) =>
      div.textContent?.includes('We couldn\'t find any projects'),
    );
    expect(emptyStateContainer).toBeTruthy();

    const refreshButton = (emptyStateContainer as HTMLElement)?.querySelector('button');
    expect(refreshButton).toBeFalsy();
  });

  it('should center the empty state vertically and horizontally', () => {
    fixture.detectChanges();

    const emptyContainer = fixture.nativeElement.querySelector('.items-center.justify-center');
    expect(emptyContainer).toBeTruthy();
  });

  it('should navigate to project view on select when default view exists', () => {
    const mockProject = {
      id: '1',
      name: 'Test Project',
      slug: 'test-project',
      description: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    (mockProjectService.getBySlug as ReturnType<typeof vi.fn>).mockReturnValue(mockProject);
    (mockViewService.getDefaultView as ReturnType<typeof vi.fn>).mockReturnValue({ id: '1', slug: 'default', name: 'Default', icon: 'grid', viewType: 'kanban' as const });

    (component as any).selectProject('test-project');

    expect(mockProjectService.setActive).toHaveBeenCalledWith(mockProject);
    expect(mockRouter.navigate).toHaveBeenCalledWith(['/', 'test-project', 'views', 'default']);
  });

  it('should navigate to project root when no default view exists', () => {
    const mockProject = {
      id: '1',
      name: 'Test Project',
      slug: 'test-project',
      description: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    (mockProjectService.getBySlug as ReturnType<typeof vi.fn>).mockReturnValue(mockProject);
    (mockViewService.getDefaultView as ReturnType<typeof vi.fn>).mockReturnValue(null);

    (component as any).selectProject('test-project');

    expect(mockRouter.navigate).toHaveBeenCalledWith(['/', 'test-project']);
  });

  it('should create project and navigate to it on successful creation', () => {
    const mockProject = {
      id: '1',
      name: 'New Project',
      slug: 'new-project',
      description: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    (mockProjectService.createProject as ReturnType<typeof vi.fn>).mockReturnValue(of(mockProject));

    (component as any).onCreateProject({ name: 'New Project', description: '' });

    expect(mockProjectService.createProject).toHaveBeenCalledWith({
      name: 'New Project',
      description: undefined,
    });
    expect(mockProjectService.loadProjects).toHaveBeenCalled();
    expect(mockRouter.navigate).toHaveBeenCalledWith(['/', 'new-project']);
  });
});

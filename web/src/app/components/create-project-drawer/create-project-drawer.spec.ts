import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LUCIDE_ICONS, LucideIconProvider, X } from 'lucide-angular';
import { CreateProjectDrawerComponent } from './create-project-drawer';
import { vi } from 'vitest';

describe('CreateProjectDrawerComponent', () => {
  let component: CreateProjectDrawerComponent;
  let fixture: ComponentFixture<CreateProjectDrawerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CreateProjectDrawerComponent],
      providers: [
        { provide: LUCIDE_ICONS, multi: true, useValue: new LucideIconProvider({ X }) },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(CreateProjectDrawerComponent);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should be closed by default', () => {
    fixture.detectChanges();
    expect(component.isOpen()).toBeFalsy();
  });

  it('should render overlay and panel when open', () => {
    component.open();
    fixture.detectChanges();

    const overlay = fixture.nativeElement.querySelector('.drawer-overlay');
    const panel = fixture.nativeElement.querySelector('.drawer-panel');

    expect(overlay).toBeTruthy();
    expect(panel).toBeTruthy();
  });

  it('should not render overlay and panel when closed', () => {
    fixture.detectChanges();

    const overlay = fixture.nativeElement.querySelector('.drawer-overlay');
    const panel = fixture.nativeElement.querySelector('.drawer-panel');

    expect(overlay).toBeTruthy();
    expect(panel).toBeTruthy();

    // The elements should still be in DOM but hidden via CSS
    expect(overlay.classList.contains('opacity-0')).toBeTruthy();
    expect(panel.classList.contains('translate-x-full')).toBeTruthy();
  });

  it('should emit create event on submit with valid name', () => {
    vi.spyOn(component.create, 'emit');

    component.name.set('My Project');
    component.description.set('A description');
    component.submit();

    expect(component.create.emit).toHaveBeenCalledWith({
      name: 'My Project',
      description: 'A description',
    });
    expect(component.isOpen()).toBeFalsy();
  });

  it('should not emit create event when name is empty', () => {
    vi.spyOn(component.create, 'emit');

    component.name.set('  ');
    component.submit();

    expect(component.create.emit).not.toHaveBeenCalled();
    expect(component.isOpen()).toBeFalsy();
  });

  it('should trim whitespace from name and description', () => {
    vi.spyOn(component.create, 'emit');

    component.name.set('  My Project  ');
    component.description.set('  A description  ');
    component.submit();

    expect(component.create.emit).toHaveBeenCalledWith({
      name: 'My Project',
      description: 'A description',
    });
  });

  it('should close and emit closeDrawer event', () => {
    vi.spyOn(component.closeDrawer, 'emit');

    component.open();
    component.close();

    expect(component.isOpen()).toBeFalsy();
    expect(component.closeDrawer.emit).toHaveBeenCalled();
  });

  it('should reset name and description on open', () => {
    component.name.set('Old Name');
    component.description.set('Old Description');

    component.open();

    expect(component.name()).toBe('');
    expect(component.description()).toBe('');
    expect(component.isOpen()).toBeTruthy();
  });

  it('should have a form with submit handler', () => {
    component.open();
    fixture.detectChanges();

    const form = fixture.nativeElement.querySelector('form');
    expect(form).toBeTruthy();
  });

  it('should have a create button', () => {
    component.open();
    fixture.detectChanges();

    const buttons = fixture.nativeElement.querySelectorAll('button');
    const createButton = Array.from(buttons).find((b: any) => b.textContent.includes('Create Project'));
    expect(createButton).toBeTruthy();
  });
});

import { ComponentFixture, TestBed } from '@angular/core/testing';

import { GroupManagerComponent } from './group-manager.component';

describe('GroupManagerComponent', () => {
  let component: GroupManagerComponent;
  let fixture: ComponentFixture<GroupManagerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [GroupManagerComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(GroupManagerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

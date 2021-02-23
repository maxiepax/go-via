import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageIsoComponent } from './manage-iso.component';

describe('ManageIsoComponent', () => {
  let component: ManageIsoComponent;
  let fixture: ComponentFixture<ManageIsoComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManageIsoComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageIsoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

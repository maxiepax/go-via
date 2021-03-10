import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageHostsComponent } from './manage-hosts.component';

describe('ManageHostsComponent', () => {
  let component: ManageHostsComponent;
  let fixture: ComponentFixture<ManageHostsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManageHostsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageHostsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

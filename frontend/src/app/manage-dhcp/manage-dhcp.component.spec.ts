import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageDhcpComponent } from './manage-dhcp.component';

describe('ManageDhcpComponent', () => {
  let component: ManageDhcpComponent;
  let fixture: ComponentFixture<ManageDhcpComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ManageDhcpComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageDhcpComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

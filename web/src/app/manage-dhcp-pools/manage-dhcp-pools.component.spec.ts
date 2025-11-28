import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageDhcpPoolsComponent } from './manage-dhcp-pools.component';

describe('ManageDhcpPoolsComponent', () => {
  let component: ManageDhcpPoolsComponent;
  let fixture: ComponentFixture<ManageDhcpPoolsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ManageDhcpPoolsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageDhcpPoolsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

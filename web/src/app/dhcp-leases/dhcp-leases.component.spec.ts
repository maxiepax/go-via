import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DhcpLeasesComponent } from './dhcp-leases.component';

describe('DhcpLeasesComponent', () => {
  let component: DhcpLeasesComponent;
  let fixture: ComponentFixture<DhcpLeasesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DhcpLeasesComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DhcpLeasesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

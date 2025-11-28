import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DhcpPoolManagerComponent } from './dhcp-pool-manager.component';

describe('DhcpPoolManagerComponent', () => {
  let component: DhcpPoolManagerComponent;
  let fixture: ComponentFixture<DhcpPoolManagerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DhcpPoolManagerComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DhcpPoolManagerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ManageImagesComponent } from './manage-images.component';

describe('ManageImagesComponent', () => {
  let component: ManageImagesComponent;
  let fixture: ComponentFixture<ManageImagesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ManageImagesComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ManageImagesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

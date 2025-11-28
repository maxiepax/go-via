import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';


@Component({
    selector: 'app-manage-images',
    templateUrl: './manage-images.component.html',
    styleUrls: ['./manage-images.component.scss'],
    standalone: false
})
export class ManageImagesComponent implements OnInit {
  images: any[] = [];




  constructor(private apiService: ApiService) {

  }

  ngOnInit(): void {
    this.apiService.getImages().subscribe((images: any) => {
      this.images = images;
    });
  }





}

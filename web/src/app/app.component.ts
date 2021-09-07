import { Component } from '@angular/core';
import { ApiService } from './api.service';


@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  version;

  constructor(private apiService: ApiService) {
    
  }

  ngOnInit(): void {
    this.apiService.getVersion().subscribe((data: any) => {
      this.version = data;
    });

  }
}

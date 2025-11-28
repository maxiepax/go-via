import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-deployments',
  templateUrl: './deployments.component.html',
  styleUrl: './deployments.component.scss',
  standalone: false
})
export class DeploymentsComponent {
    constructor( public router: Router) {}
  

}

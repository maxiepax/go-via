import { Component } from '@angular/core';
import { Router } from '@angular/router';
@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrl: './settings.component.scss',
  standalone: false
})
export class SettingsComponent {
    constructor( public router: Router) {}
}

import { Component, OnInit, Renderer2 } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../auth.service';
import { ApiService } from '../api.service';
import { FormsModule } from '@angular/forms';
import { ClarityModule } from '@clr/angular';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss'],
  standalone: true,
  imports: [FormsModule, ClarityModule, CommonModule]
})
export class LoginComponent implements OnInit {
  username: string;
  password: string;
  errorMessage: string;

  form = {
    type: 'local',
    username: '',
    password: '',
    rememberMe: false,
  };

  constructor(
    private authService: AuthService,
    private router: Router,
    private api: ApiService,
    private renderer: Renderer2
  ) {}

  ngOnInit(): void {
    this.api.getThemeImage().subscribe({
      next: (blob) => {
        if (blob && blob.size > 0) {
          const reader = new FileReader();
          reader.onload = () => {
            this.setBackground(reader.result as string);
          };
          reader.readAsDataURL(blob);
        } else {
          this.setBackground('/assets/background.png');
        }
      },
      error: () => {
        this.setBackground('/assets/background.png');
      }
    });
  }

  setBackground(url: string) {
    const el = document.querySelector('.login-wrapper') as HTMLElement;
    if (el) {
      this.renderer.setStyle(el, 'background-image', `url('${url}')`);
    }
  }

  login() {
    this.authService.login(this.form.username, this.form.password).subscribe(
      (resp: any) => {
        console.log('Login successful');
        localStorage.setItem('username', this.form.username);
        this.router.navigate(['/']);
      },
      (error) => {
        console.log(error);
        this.errorMessage = "Invalid username or password";
        this.router.navigate(['/login']);
      }
    );
  }
}


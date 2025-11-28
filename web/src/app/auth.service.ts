import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { HttpClient } from '@angular/common/http';

import { BehaviorSubject, of } from 'rxjs';
import { map } from 'rxjs/operators';


@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private isAuthenticated = false;
  private usernameSubject = new BehaviorSubject<string>(localStorage.getItem('username') || '');
  username$ = this.usernameSubject.asObservable();


  constructor(private router: Router, private httpClient: HttpClient) {}

  login(username: string, password: string) {

    const body = { username, password };

    var resp = this.httpClient.post(
      'https://' + window.location.host + '/v1/login',
      body
    );

    this.usernameSubject.next(username);
    localStorage.setItem('username', username);
    this.isAuthenticated = true;
    console.log(this.isAuthenticated);



    return resp
  }

  logout(): void {
    this.isAuthenticated = false;
    this.usernameSubject.next('');
    localStorage.removeItem('username');
    this.router.navigate(['/login']);
  }

  isLoggedIn(): boolean {
    console.log(this.isAuthenticated);
    // return true; // use this for local dev
    return this.isAuthenticated;
  }
}
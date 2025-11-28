import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import {
  UntypedFormBuilder,
  FormControl,
  UntypedFormGroup,
  Validators
} from '@angular/forms';

@Component({
    selector: 'app-manage-users',
    templateUrl: './manage-users.component.html',
    styleUrls: ['./manage-users.component.scss'],
    standalone: false
})
export class ManageUsersComponent implements OnInit {
  user;
  users;
  errors;
  form: UntypedFormGroup;
  showUserModalMode = "";

  constructor(
    private apiService: ApiService,
    private formBuilder: UntypedFormBuilder
  ) {
    this.form = this.formBuilder.group({
      username: ['', [Validators.required]],
      password: ['', [Validators.required]],
      email: ['', [Validators.required]],
      comment: ['', [Validators.required]],
    });
   }

  ngOnInit(): void {
    this.apiService.getUsers().subscribe((data: any) => {
      this.users = data;
    });
  }

  showUserModal(mode, id=null) {
    this.showUserModalMode = mode;
    if (mode === "edit") {
    this.user = this.users.find(user => user.id === id);
    this.form.patchValue({
      ...this.user,
      password: undefined,
    });
    }
    if (mode === "add") {
      this.form.reset();
    }
  }




  submit() {
    const data = {
      ...this.form.value,
    };

    this.apiService.addUser(data).subscribe((resp: any) => {
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.users.push(resp);
        this.form.reset();
      }
    });

    this.showUserModalMode = '';
  }

  remove(id) {
    this.apiService.deleteUser(id).subscribe((data: any) => {
      this.users = this.users.filter(item => item.id !== id);
    });
  }

  update() {
    const data = {
      ...this.form.value,
    };

    this.apiService.updateUser(this.user.id, data).subscribe((resp: any) => {
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.users = this.users.filter(item => item.id !== resp.id);
        this.users.push(resp);
        this.showUserModalMode = '';
      }
    });
  }

}

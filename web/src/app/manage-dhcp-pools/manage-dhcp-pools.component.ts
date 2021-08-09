import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import {
  FormBuilder,
  FormControl,
  FormGroup,
  Validators
} from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@Component({
  selector: 'app-manage-dhcp-pools',
  templateUrl: './manage-dhcp-pools.component.html',
  styleUrls: ['./manage-dhcp-pools.component.scss']
})
export class ManageDhcpPoolsComponent implements OnInit {
  pool;
  pools;
  errors;
  form: FormGroup;
  showPoolModalMode = "";

  constructor(
    private apiService: ApiService,
    private formBuilder: FormBuilder
  ) {
    this.form = this.formBuilder.group({
      net_address: ['', [Validators.required]],
      netmask: ['', [Validators.required]],
      name: ['', [Validators.required]],
      start_address: ['', [Validators.required]],
      end_address: ['', [Validators.required]],
      gateway: ['', [Validators.required]],
    });
  }

  ngOnInit(): void {
    this.apiService.getPools().subscribe((data: any) => {
      this.pools = data;
    });
  }


  showPoolModal(mode, id=null) {
    this.showPoolModalMode = mode;
    if (mode === "edit") {
    this.pool = this.pools.find(pool => pool.id === id);
    this.form.patchValue({
      ...this.pool,
    });
    }
    if (mode === "add") {
      this.form.reset();
    }
  }

  submit() {
    const data = {
      ...this.form.value,
      only_serve_reimage: true,
      lease_time: 7000,
    };

    this.apiService.addPool(data).subscribe((resp: any) => {
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.pools.push(resp);
        this.form.reset();
      }
    });

    this.showPoolModalMode = '';
  }

  remove(id) {
    this.apiService.deletePool(id).subscribe((data: any) => {
      this.pools = this.pools.filter(item => item.id !== id);
    });
  }

  updatePool() {
    const data = {
      ...this.form.value,
      only_serve_reimage: true,
      lease_time: 7000,
    };

    this.apiService.updatePool(this.pool.id, data).subscribe((resp: any) => {
      console.log(resp);
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.pools = this.pools.filter(item => item.id !== resp.id);
        this.pools.push(resp);
        this.showPoolModalMode = '';
      }
    });
  }
}

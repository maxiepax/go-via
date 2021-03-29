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
  showPoolFormModal = false;
  editPoolFormModal = false;

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
      dns: ['', [Validators.required]]
    });
  }

  ngOnInit(): void {
    this.apiService.getPools().subscribe((data: any) => {
      this.pools = data;
    });
  }

  addPool() {
    this.showPoolFormModal = true;
  }

  editPool(id) {
    this.editPoolFormModal = true;
    this.pool = this.pools.find(pool => pool.id === id);
    this.form.patchValue({
      ...this.pool,
      dns: (this.pool.dns || []).join(', ')
    });
  }

  submit() {
    const data = {
      ...this.form.value,
      only_serve_reserved: true,
      lease_time: 7000,
      dns: this.form.value.dns.split(',').map(a => a.trim())
    };

    this.apiService.addPool(data).subscribe((resp: any) => {
      console.log(resp);
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.pools.push(resp);
        this.form.reset();
      }
    });
  }

  remove(id) {
    this.apiService.deletePool(id).subscribe((data: any) => {
      this.pools = this.pools.filter(item => item.id !== id);
    });
  }

  updatePool() {
    console.log('test');
    const data = {
      ...this.form.value,
      only_serve_reserved: true,
      lease_time: 7000,
      dns: this.form.value.dns.split(',').map(a => a.trim())
    };

    this.apiService.updatePool(this.pool.id, data).subscribe((resp: any) => {
      console.log(resp);
      if (resp.error) {
        this.errors = resp.error;
      }
      if (resp) {
        this.pools = this.pools.filter(item => item.id !== resp.id);
        this.pools.push(resp);
        this.editPoolFormModal = false;
      }
    });
  }
}

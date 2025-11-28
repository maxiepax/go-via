import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ApiService } from '../../api.service';
@Component({
  selector: 'app-dhcp-pool-manager',
  templateUrl: './dhcp-pool-manager.component.html',
  styleUrls: ['./dhcp-pool-manager.component.scss'],
  standalone: false
})
export class DhcpPoolManagerComponent implements OnInit {
  @Input() pools: any[] = [];
  @Output() poolsChange = new EventEmitter<any[]>();
  errors: string[] = [];
  form: FormGroup;
  showPoolModalMode = false;
  editIndex: number | null = null;

  editMode: boolean = false;

  constructor(
    private fb: FormBuilder,     
    private apiService: ApiService
  ) {
    this.form = this.fb.group({
      net_address: ['', Validators.required],
      netmask: ['', Validators.required],
      name: ['', Validators.required],
      start_address: ['', Validators.required],
      end_address: ['', Validators.required],
      gateway: ['', Validators.required],
    });
  }

  ngOnInit(): void {}

  showPoolModal(mode: 'add' | 'edit', id?: number) {
    console.log('showPoolModal called with mode:', mode, 'id:', id);
    console.log('Current pools:', this.pools);

    if (mode === 'add') {
      this.editMode = false;
      this.openAddModal();
    } else if (mode === 'edit' && typeof id === 'number') {
      this.editMode = true;
      const index = this.pools.findIndex(pool => pool.id === id);
      console.log('Found pool index:', index, 'for id:', id);
      if (index !== -1) {
        console.log('Pool to edit:', this.pools[index]);
        this.openEditModal(index);
      } else {
        console.error('Pool not found with id:', id);
        this.errors = ['Pool not found'];
      }
    }
  }

  openAddModal() {
    this.form.reset();
    this.editIndex = null;
    this.showPoolModalMode = true;
  }

  openEditModal(index: number) {
    console.log('openEditModal called with index:', index);
    console.log('Pool at index:', this.pools[index]);
    
    this.editIndex = index;
    this.form.patchValue(this.pools[index]);
    this.showPoolModalMode = true;
    this.errors = []; // Clear any previous errors
  }

  submit() {
    if (this.form.invalid) return;

    const data = {
      ...this.form.value,
      only_serve_reimage: true,
      lease_time: 7000,
    };

    console.log('Submitting form data:', data);


    if (this.editMode && this.editIndex !== null) {
      const poolId = this.pools[this.editIndex].id;
      console.log('Edit mode - Pool ID:', poolId, 'Edit Index:', this.editIndex, 'Pool:', this.pools[this.editIndex]);
      
      if (!poolId || poolId === 0) {
        console.error('Invalid pool ID for update:', poolId);
        this.errors = ['Invalid pool ID for update'];
        return;
      }
      
      this.apiService.updatePool(poolId, data).subscribe((resp: any) => {
        if (resp.error) {
          this.errors = resp.error;
        }
        if (resp) {
          this.pools[this.editIndex!] = resp;
          this.form.reset();
          this.poolsChange.emit(this.pools);
        }
      });
    } else {
      console.log('Add mode - Creating new pool');
      this.apiService.addPool(data).subscribe((resp: any) => {
        if (resp.error) {
          this.errors = resp.error;
        }
        if (resp) {
          this.pools.push(resp);
          this.form.reset();
          this.poolsChange.emit(this.pools);
        }
      });
    }
    this.showPoolModalMode = false;
    this.editIndex = null;
  }

  delete(index: number) {
    this.pools.splice(index, 1);
    this.poolsChange.emit(this.pools);
  }
}

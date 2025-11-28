import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { FormBuilder, UntypedFormGroup, Validators } from '@angular/forms';
import { Observable } from 'rxjs';
import { ApiService } from 'src/app/api.service';

@Component({
  selector: 'app-group-manager',
  templateUrl: './group-manager.component.html',
  styleUrls: ['./group-manager.component.scss'],
  standalone: false
})
export class GroupManagerComponent implements OnInit {
  @Input() group: any = {};
  @Input() mode: 'add' | 'edit' = 'add';
  @Input() errors: any = null;
  @Input() pools: any[] = [];    // <-- Add this line
  @Input() images: any[] = [];   // <-- Add this line
  @Output() groupChange = new EventEmitter<any>();
  @Output() submit = new EventEmitter<any>();

  groupForm: UntypedFormGroup;

  advanced = false;

  constructor(private fb: FormBuilder, private apiService: ApiService,) {
    this.groupForm = this.fb.group({
      name: ['', Validators.required],
      pool_id: [''],
      image_id: [''],
      password: [''],
      bootdisk: [''],
      vlan: [''],
      callbackurl: [''],
      dns: [''],
      ntp: [''],
      syslog: [''],
      ssh: [false],
      erasedisks: [false],
      allowlegacycpu: [false],
      certificate: [false],
      createvmfs: [false],
      ks: ['']
    });
  }

  ngOnInit() {
      this.apiService.getImages().subscribe((images: any) => {
      this.images = images;
    });
    this.apiService.getPools().subscribe((pools: any) => {
      this.pools = pools;
    });
  }

  ngOnChanges() {
    if (this.group && this.groupForm) {
      this.groupForm.patchValue(this.group);
    }
  }

  onSubmit() {
    if (this.groupForm.valid) {
      
    }
  }

  submitGroup(): Observable<any> {
    
    const data = {
      ...this.groupForm.value,
      image_id: parseInt(this.groupForm.value.image_id),
      pool_id: parseInt(this.groupForm.value.pool_id),
    }

    console.log('Submitting group data:', data);

    let json_pc: any = {}
    if (data.ssh) {
      json_pc.ssh = true;
    }
    if (data.erasedisks) {
      json_pc.erasedisks = true;
    }
    if (data.allowlegacycpu) {
      json_pc.allowlegacycpu = data.allowlegacycpu;
    }
    if (data.certificate) {
      json_pc.certificate = data.certificate;
    }
    if (data.createvmfs) {
      json_pc.createvmfs = data.createvmfs;
    }
    if (data.ks) {
      data.ks = btoa(data.ks)
    }

    data.options = json_pc;
    delete data.ssh;
    delete data.erasedisks;
    delete data.allowlegacycpu;
    delete data.certificate;
    delete data.createvmfs;

    return this.apiService.addGroup(data);
  }
}

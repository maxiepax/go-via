import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { cloudIcon, ClarityIcons } from '@cds/core/icon';
import '@cds/core/icon/register.js';
import '@cds/core/accordion/register.js';
import '@cds/core/alert/register.js';
import '@cds/core/button/register.js';
import '@cds/core/checkbox/register.js';
import '@cds/core/datalist/register.js';
import '@cds/core/file/register.js';
import '@cds/core/forms/register.js';
import '@cds/core/input/register.js';
import '@cds/core/password/register.js';
import '@cds/core/radio/register.js';
import '@cds/core/range/register.js';
import '@cds/core/search/register.js';
import '@cds/core/select/register.js';
import '@cds/core/textarea/register.js';
import '@cds/core/time/register.js';
import '@cds/core/toggle/register.js';

@Component({
  selector: 'app-manage-groups',
  templateUrl: './manage-groups.component.html',
  styleUrls: ['./manage-groups.component.scss']
})
export class ManageGroupsComponent implements OnInit {
	hosts;
	images;
	errors;
  groups;
	form: FormGroup;
  groupid = null;
  showHostFormModal = false;
  showGroupFormModal = false;

  constructor(private apiService: ApiService, private formBuilder: FormBuilder) {
  	this.form = this.formBuilder.group({
		fqdn: ['', [Validators.required]],
		ip: ['', [Validators.required]],
		mac: ['', [Validators.required]],
	    group_id: ['', [Validators.required]],
	});
  }


  ngOnInit(): void {
    this.apiService.getGroups().subscribe((groups:any)=>{	
      this.apiService.getHosts().subscribe((hosts:any)=>{
        this.groups = groups.map(item => {
          item.hosts = hosts.filter(host => host.group_id === item.id)
          return item
        });
        //console.log(groups);
      });
    });
  }

  addGroup() {
    
  }

  addHostToGroup(id) {
    this.showHostFormModal = true;
    this.groupid = id;
  }


  submit_host() {
    const data={
      ...this.form.value,
      //active:true,
      hostname: this.form.value.fqdn.split(".")[0],
      domain: this.form.value.fqdn.split(".").slice(1).join('.'),
      group_id: parseInt(this.groupid),
    }
  
    this.apiService.addHost(data).subscribe((data:any)=>{
      
      if (data.id) {
        this.groups.find(group => group.id === data.group_id).hosts.push(data)
        this.form.reset();
        this.showHostFormModal = false;
      }
    },(data:any)=>{
      console.log(data);
      if (data.status) {
        this.errors = data.error;
      } else {
        this.errors = [data.message];
      }
    });
  }
  
  submit_group() {
    const data={
      ...this.form.value,
      //active:true,
      hostname: this.form.value.fqdn.split(".")[0],
      domain: this.form.value.fqdn.split(".").slice(1).join('.'),
      group_id: parseInt(this.groupid),
    }
  
    this.apiService.addHost(data).subscribe((data:any)=>{
      
      if (data.id) {
        this.groups.find(group => group.id === data.group_id).hosts.push(data)
        this.form.reset();
        this.showHostFormModal = false;
      }
    },(data:any)=>{
      console.log(data);
      if (data.status) {
        this.errors = data.error;
      } else {
        this.errors = [data.message];
      }
    });
  }

    remove(id) {
      console.log(id);
      this.apiService.deleteHost(id).subscribe((data:any)=>{
      /*
      console.log("return data");
          console.log(data);
      this.groups.hosts = this.groups.hosts.filter(item => item.id !== id);
      */
      this.groups = this.groups.map(item => {
        item.hosts = item.hosts.filter(host => id !== host.id)
        return item
      });
      });
    }
}

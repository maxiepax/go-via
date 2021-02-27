import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

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
      });
    });
  }

  submit() {
    const data={
      ...this.form.value,
      //active:true,
      hostname: this.form.value.fqdn.split(".")[0],
      domain: this.form.value.fqdn.split(".").slice(1).join('.'),
      group_id: parseInt(this.form.value.group_id),
    }
  
    this.apiService.addHost(data).subscribe((data:any)=>{
      if (data.error) {
        this.errors = data.error;
      }
      if (data.host) {
        this.hosts.push(data);
        this.form.reset();
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

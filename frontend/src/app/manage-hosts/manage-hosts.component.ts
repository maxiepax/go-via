import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@Component({
  selector: 'app-manage-hosts',
  templateUrl: './manage-hosts.component.html',
  styleUrls: ['./manage-hosts.component.scss']
})
export class ManageHostsComponent implements OnInit {
	hosts;
	images;
	errors;
	form: FormGroup;

//  constructor(private apiService: ApiService) { }

  constructor(private apiService: ApiService, private formBuilder: FormBuilder) {
  	this.form = this.formBuilder.group({
		fqdn: ['', [Validators.required]],
		ip_address: ['', [Validators.required]],
		mac: ['', [Validators.required]],
	    password: ['', [Validators.required]],
	    image_id: ['', [Validators.required]],
	});
  }

  ngOnInit(): void {
	this.apiService.getHosts().subscribe((data:any)=>{
		console.log(data);
		this.hosts = data.hosts;
	});

	this.apiService.getImages().subscribe((data:any)=>{
		console.log(data);
		this.images = data.images;
	});
  }

  submit() {
	const data={
		...this.form.value,
		active:true,
		hostname: this.form.value.fqdn.split(".")[0],
	}
	this.apiService.addHost(data).subscribe((data:any)=>{
		console.log(data);
		if (data.error) {
			this.errors = data.error;
		}
		if (data.host) {
			this.hosts.push(data.host);
			this.form.reset();
		}
	});

  }
  remove() {
	console.log('remove');
  }

}


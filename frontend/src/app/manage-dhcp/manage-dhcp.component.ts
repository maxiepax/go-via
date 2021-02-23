import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@Component({
  selector: 'app-manage-dhcp',
  templateUrl: './manage-dhcp.component.html',
  styleUrls: ['./manage-dhcp.component.scss']
})
export class ManageDhcpComponent implements OnInit {
	pools;
	errors;
	form: FormGroup;

  constructor(private apiService: ApiService, private formBuilder: FormBuilder) { 
	this.form = this.formBuilder.group({
		net_address: ['', [Validators.required]],
		netmask: ['', [Validators.required]],
		name: ['', [Validators.required]],
		start_address: ['', [Validators.required]],
		end_address: ['', [Validators.required]],
		gateway: ['', [Validators.required]],
		dns: ['', [Validators.required]],
		image: ['', [Validators.required]],
	});
  }

  ngOnInit(): void {
	this.apiService.getPools().subscribe((data:any)=>{
		console.log(data);
		this.pools = data.pools;
	});

  }

  submit() {

	const data={ 
		...this.form.value,
		only_serve_reserved: true,
	}

	this.apiService.addPool(data).subscribe((data:any)=>{
		console.log(data);
		if (data.error) {
			this.errors = data.error;
		}
		if (data.pool) {
			this.pools.push(data.pool);
			this.form.reset();
		}
	});
  }

}

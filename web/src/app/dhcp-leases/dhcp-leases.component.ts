import { Component, OnInit } from '@angular/core';
import { ApiService } from '../api.service';

@Component({
  selector: 'app-dhcp-leases',
  templateUrl: './dhcp-leases.component.html',
  styleUrls: ['./dhcp-leases.component.scss'],
  standalone: false
})




export class DhcpLeasesComponent implements OnInit {
  groupedLeases: { domain: string; leases: Lease[] }[] = [];

  constructor(private apiService: ApiService) {}

  ngOnInit(): void {
    this.apiService.getFormattedAddresses().subscribe((data: Lease[]) => {
      const grouped = data.reduce((acc, lease) => {
        const domain = lease.domain;
        if (!acc[domain]) {
          acc[domain] = [];
        }
        acc[domain].push(lease);
        return acc;
      }, {} as { [key: string]: Lease[] });

      this.groupedLeases = Object.keys(grouped).map((domain) => ({
        domain,
        leases: grouped[domain],
      }));

      console.log('Grouped Leases:', this.groupedLeases); // Debug statement
    });
  }
}


export interface Lease {
  ip: string;
  hostname: string;
  domain: string;
  expires_at: string;
}
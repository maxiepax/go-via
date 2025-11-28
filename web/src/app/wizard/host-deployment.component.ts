import { CommonModule } from '@angular/common';
import { Component, ElementRef, OnInit, ViewChild, AfterViewInit, AfterViewChecked } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../api.service';
import { ClrFormsModule, ClrWizard, ClrWizardModule, ClarityModule } from '@clr/angular';
import { DhcpPoolManagerComponent } from '../components/dhcp-pool-manager/dhcp-pool-manager.component';
import { GroupManagerComponent } from '../components/group-manager/group-manager.component';

@Component({
  selector: 'app-host-deployment',
  templateUrl: './host-deployment.component.html',
  styleUrls: ['./host-deployment.component.scss'],
  standalone: false,

})
export class HostDeploymentComponent implements OnInit {
  @ViewChild('wizard', { static: true }) wizard: ClrWizard | undefined;
  @ViewChild('fileInput', { static: false }) fileInput!: ElementRef;
  @ViewChild(DhcpPoolManagerComponent) poolManager?: DhcpPoolManagerComponent;
  @ViewChild(GroupManagerComponent) groupManager?: GroupManagerComponent;
  open = false;
  model: any;
  loadingFlag = false;
  errorFlag = false;
  pools: any[] = [];
  selectedHosts: Host[] = []; // Use the Host type for selected hosts
  showAddGroupForm = false;
  groups: any[] = [];
  selectedGroup: any = null;
  images: any[] = [];


  // Add this property for the modal open state
  dhcpPoolModalOpen = false;
  configRetrievalSuccessFlag: boolean;

  constructor(private apiService: ApiService) { }

  ngOnInit() {
    this.model = {
      hosts: [] as Host[], // List of all hosts
      errors: [],
      forceReset: true,
      useSameCreds: false,
      favoriteColor: '',
      luckyNumber: '',
      flavorOfIceCream: '',
      selectVendor: '',
      iloPort: 443,
      setIloPort: false,
      username: '',
      password: '',
      selectedVendor: '',
      domain: '', // Add domain property
    };

    // Fetch groups from the API
    this.apiService.getGroups().subscribe((groups: any) => {
      this.groups = groups;
      // Auto-select the first group if only one exists
      this.setDefaultGroupSelection();
    });

    // Fetch Images from the API
    this.apiService.getImages().subscribe((images: any) => {
      this.images = images;
    });
  }

  setDefaultGroupSelection(): void {
    if (this.groups.length > 0) {
      this.selectedGroup = this.groups[0].id.toString();
      this.model.selectedGroup = this.groups[0];
      console.log('Auto-selected default group:', this.model.selectedGroup);
    }
  }

  selectVendor(vendor: string): void {
    this.model.selectedVendor = vendor;
    console.log('Selected vendor:', vendor);
  }

  getApiFlavourFromVendor(): string {
    switch (this.model.selectedVendor) {
      case "HPE": return 'redfish';
      default: return 'redfish';
    }
  }


  doCancel(): void {

    this.wizard?.reset(); // Explicitly reset the wizard
    this.open = false; // Close the wizard
  }


  async addToGroup(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.selectedGroup) {
        this.model.errors('No group selected.');
        this.model.errors.push(new Error('No group selected.'));
        reject(new Error('No group selected.'));
        return;
      }

      if (this.selectedHosts.length === 0) {
        console.error('No hosts selected.');
        this.model.errors.push(new Error('No hosts selected.'));
        reject(new Error('No hosts selected.'));
        return;
      }



      // Prepare all host data for the API calls
      const hostRequests = this.selectedHosts.map((host) => {
        const hostData = {
          fqdn: host.fqdn,
          // ip: host.selectedIface?.ipAddress, // TODO: what ip is really needed here?
          ip: host.selectedIface.ipAddress, // Add other properties you need
          mac: host.selectedIface?.macAddress, // If required
          hostname: host.fqdn.split(".")[0],
          domain: host.fqdn.split(".").slice(1).join('.'),
          group_id: parseInt(this.selectedGroup || '', 10),
          pool_id: this.model.selectedGroup.id || 0,
          ilo_ip: host.iloIpAddr,
          ilo_fqdn: host.iloFqdn || '', // Use fqdn if available
          ilo_user: host.username,
          ilo_password: host.password,
          ilo_port: String(host.iloPort),
          ilo_api_flavour: this.getApiFlavourFromVendor(),
        };

        console.log('Prepared host data for API:', hostData);
        // Return the API call as a promise
        return this.addHostPromise(hostData);
      });

      // Wait for all API calls to complete
      Promise.all(hostRequests)
        .then((responses) => {
          console.log('All hosts added successfully:', responses);
          resolve(); // Resolve the promise
          this.wizard?.next();
        })
        .catch((error) => {
          console.error('Error adding hosts:', error);
          this.model.errors.push({
            subject: 'Error adding hosts',
            message: error.error.error_message || error,
          });
          this.errorFlag = true; // Set error flag
          reject(error); // Reject the promise
        });
    });
  }

  newHostInput: any = { iloIpAddr: '', username: '', password: '', iloFqdn: '' }; // Initialize new host input
  addHostVoid(): void {
    // Ensure the hosts array exists
    if (!Array.isArray(this.model.hosts)) {
      this.model.hosts = [];
    }

    if (this.model.useSameCreds) {
      this.newHostInput.username = this.model.username;
      this.newHostInput.password = this.model.password;
    }
    console.log(this.newHostInput);

    // Add the new host to the hosts array
    this.model.hosts.push({
      iloIpAddr: this.newHostInput.iloIpAddr,
      iloFqdn: this.newHostInput.iloFqdn || '', // Use iloFqdn if provided
      iloPort: this.model.iloPort,
      username: this.newHostInput.username,
      password: this.newHostInput.password,
    });

    // Reset the input fields for the next host
    this.newHostInput = { iloIpAddr: '', username: '', password: '' };

    console.log(this.model.hosts);
  }


  addHostPromise(data: any): Promise<any> {
    return new Promise((resolve, reject) => {
      if (data.ks) {
        data.ks = btoa(data.ks);
      }

      this.apiService.addHost(data).subscribe({
        next: (response: any) => {
          if (response.id) {
            const group = this.groups.find((group) => group.id === data.group_id);
            if (group) {
              group.hosts = [...(group.hosts || []), response];
            }
            resolve(response); // Resolve the promise with the response
          } else {
            reject(new Error('Host creation failed: No ID returned.'));
          }
        },
        error: (error: any) => {
          console.error('Error adding host:', error.error.error_message || error);
          reject(error); // Reject the promise with the error
        },
      });
    });
  }

  editHost(index: number): void {
    const host = this.model.hosts[index];
    this.newHostInput = { ...host }; // Populate input fields with the selected host's data
    this.removeHost(index); // Remove the host temporarily to allow editing
  }

  removeHost(index: number): void {
    this.model.hosts.splice(index, 1); // Remove the host at the specified index
  }

  goBack(): void {
    this.wizard.previous();
  }
  doReset(): void {
    this.wizard?.reset();
    this.model.forceReset = true;
    this.model.favoriteColor = '';
    this.model.luckyNumber = '';
    this.model.flavorOfIceCream = '';
    this.loadingFlag = false;
    this.errorFlag = false;
    this.model.hosts = [];
    this.model.errors = [];
    this.model.selectedVendor = '';
    this.model.useSameCreds = false;
    this.model.iloPort = 443;
    this.model.setIloPort = false;
  }

  // IMPORT
  // import funcs to handle bulk import of hosts

  triggerFileInput(): void {
    if (this.fileInput) {
      // Reset the file input value to ensure the change event is triggered
      this.fileInput.nativeElement.value = '';
      this.fileInput.nativeElement.click();
    } else {
      console.error('fileInput is not available. Ensure the wizard is open and the element is rendered.');
    }
  }

  importHostsFromCSVToGroup(event: Event): void {
    console.log('Importing hosts from CSV to group');
    const target = event.target as HTMLInputElement;

    if (target.files && target.files.length > 0) {
      const file = target.files[0];
      console.log('Selected file:', file.name);

      // Call readCSVFile and handle the Promise
      this.readCSVFile(file)
        .then((parsedData) => {
          console.log('Parsed Host Data:', parsedData);

          parsedData.forEach((host) => {
            // Add the new host to the hosts array
            this.model.hosts.push({
              iloIpAddr: host.iloIpAddr,
              username: host.username,
              password: host.password,
              fqdn: host.fqdn,
            });

            console.log(this.model.hosts);
          });
        })
        .catch((error) => {
          console.error('Error:', error);
          // Handle error, such as showing a notification to the user
        });
    }
  }

  readCSVFile(file: File): Promise<any[]> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();

      reader.onload = (e) => {
        const csvContent = e.target?.result as string;

        try {
          const parsedData = this.parseCSV(csvContent);
          resolve(parsedData); // Resolve with parsed data
        } catch (error) {
          reject('Error parsing CSV: ' + error); // Reject if there's an error parsing
        }
      };

      reader.onerror = (error) => {
        reject('Error reading file: ' + error); // Reject if there's an error reading the file
      };

      reader.readAsText(file);
    });
  }
  // Parses CSV string into an array of objects matching the form structure
  parseCSV(csv: string): any[] {
    const lines = csv.split('\n').map(line => line.trim()).filter(line => line);
    const headers = lines[0].split(',').map(h => h.trim()); // Extract headers

    const data: any[] = [];

    for (let i = 1; i < lines.length; i++) {
      const values = lines[i].split(',').map(v => v.trim());
      if (values.length === headers.length) {
        const formGroup = {};

        // Dynamically map values to the formGroup based on the header fields
        headers.forEach((header, index) => {
          let value = values[index] || ''; // Default to empty string if no value
          formGroup[header] = value;

        });

        // Push the formGroup into the data array
        data.push(formGroup);
      }
    }
    return data;
  }

  // END IMPORT

  // VALIDATE

  async onCommit(): Promise<void> {
    this.loadingFlag = true;
    this.errorFlag = false;

    console.log(this.model.hosts);

    // Clear previous errors
    this.model.errors = [];

    // Ensure a vendor has been chosen
    if (!this.model.selectedVendor) {
      this.model.errors.push({
        subject: 'Vendor',
        message: 'Please select a vendor before proceeding.',
      });
    }

    for (const host of this.model.hosts) {

      await this.validateHost(host); // Wait for each host validation to complete
    }

    console.log('Validation done');
    console.log(this.model.errors);

    this.loadingFlag = false;

    // Set error flag if there are errors
    if (this.model.errors.length > 0) {
      this.errorFlag = true;
    } else {
      this.wizard?.next();
    }
  }

  validateHost(host: any): Promise<void> {
    return new Promise((resolve) => {
      console.log('Validating host ilo ip addr:', host.iloIpAddr);

      this.apiService.checkILOM(host.iloIpAddr, this.model.iloPort).subscribe({
        next: (resp: any) => {
          console.log('Response:', resp);
          if (resp.error) {
            this.model.errors.push({
              subject: host.iloIpAddr,
              message: resp.error,
            });
          }
          resolve(); // Resolve the promise after validation is complete
        },
        error: (err: any) => {
          console.error('Error:', err.error.message ?? err);
          this.model.errors.push({
            subject: host.iloIpAddr,
            message: err.error.message,
          });
          resolve(); // Resolve even if there's an error
        }
      });
    });
  }

  // Toggle selection for a single host
  toggleHostSelection(host: any, event: Event): void {
    const isChecked = (event.target as HTMLInputElement).checked;
    host.selected = isChecked;

    if (isChecked) {
      this.selectedHosts.push(host); // Add to selected hosts
    } else {
      this.selectedHosts = this.selectedHosts.filter(
        (h) => h.iloIpAddr !== host.iloIpAddr
      ); // Remove from selected hosts
    }

    console.log('Selected Hosts:', this.selectedHosts);
  }

  // Toggle selection for all hosts
  toggleSelectAll(event: Event): void {
    const isChecked = (event.target as HTMLInputElement).checked;

    this.model.hosts.forEach((host: any) => {
      host.selected = isChecked;
    });

    this.selectedHosts = isChecked ? [...this.model.hosts] : [];
    console.log('Selected Hosts:', this.selectedHosts);
  }

  // Process selected hosts by calling the backend API
  processSelectedHosts(): void {
    this.loadingFlag = true; // Activate loading spinner
    this.errorFlag = false; // Reset error flag
    let hasError = false; // Track if any errors occur

    if (this.selectedHosts.length === 0) {
      console.warn('No hosts selected.');
      this.loadingFlag = false; // Deactivate loading spinner
      return;
    }

    const apiCalls = this.selectedHosts.map((host) =>
      this.apiService.getHostConfig(host.iloIpAddr, this.model.iloPort, this.getApiFlavourFromVendor(), host.username, host.password).toPromise()
        .then((response) => {
          console.log(`Processed host ${host.iloIpAddr}:`, response);
          if (response.error) {
            hasError = true;
            this.model.errors.push({
              subject: host.iloIpAddr,
              message: response.error,
            });
          } else {
            console.log(`Host ${host.iloIpAddr} processed successfully. received Hostconfig:`, response.hostConfig);
          }
          // assume response.hostConfig is an array
          if (!host.ifaceConfig) {
            host.ifaceConfig = []; // Initialize ifaceConfig if undefined
          }
          for (const ifaceConfig of response.hostConfig) {
            console.log(`received Hostconfig:`, ifaceConfig);
            // Push the host configuration into the ifaceConfig array
            host.ifaceConfig.push(ifaceConfig);
          }
        })
        .catch((error) => {
          hasError = true;
          console.error(`Error processing host ${host.iloIpAddr}:`, error);
          this.model.errors.push({
            subject: host.iloIpAddr,
            message: error.error?.message || 'Unknown error occurred',
          });
        })
    );

    // Wait for all API calls to complete
    Promise.all(apiCalls).then(() => {
      this.loadingFlag = false; // Deactivate loading spinner

      if (hasError) {
        this.errorFlag = true; // Display error message
      } else {
        this.configRetrievalSuccessFlag = true;
      }
    });
  }

  onIloPortToggle(event: boolean): void {
    this.model.setIloPort = event;
    if (!event) {
      this.model.iloPort = 443; // Reset to default port if unchecked
    }
  }

  onIfaceSelectionChange(host: Host, ifaceConfig: IfaceConfig): void {
    console.log(`Selected interface for host ${host.iloIpAddr}:`, ifaceConfig);
    host.selectedIface = ifaceConfig; // Update the selected interface for the host
  }

  onGroupSelectionChange(event: Event): void {
    const selectedGroupId = (event.target as HTMLSelectElement).value;
    this.selectedGroup = selectedGroupId;
    console.log('Selected Group ID:', this.selectedGroup);

    // Find the selected group by ID
    const selectedGroup = this.groups.find((group) => group.id === parseInt(this.selectedGroup || '', 10));
    if (selectedGroup) {
      console.log('Selected Group Details:', selectedGroup);
      this.model.selectedGroup = selectedGroup; // Update the model with the selected group name
    } else {
      console.warn('No group found with the selected ID.');
      this.model.selectedGroup = null; // Reset if no group is found
    }
  }

  onGroupCommit(): void {
    console.log('=== Group Commit Function Triggered ===');
    
    // Update selectedGroup variable if not already set
    if (!this.selectedGroup && this.groups.length > 0) {
      this.selectedGroup = this.groups[0].id.toString();
      console.log('Auto-selected first group as default:', this.selectedGroup);
    }

    // Find the selected group by ID
    const selectedGroup = this.groups.find((group) => group.id === parseInt(this.selectedGroup || '', 10));
    
    if (selectedGroup) {
      // Update the model with the selected group
      this.model.selectedGroup = selectedGroup;
      

      // Check if pool is present and log pool information
      if (selectedGroup.pool_id) {
        console.log('Pool ID:', selectedGroup.pool_id);
        

      } else {
        console.warn('No pool associated with this group');
      }
      
      // Log complete group object for debugging
      

    } else {
      console.error('No group found with the selected ID:', this.selectedGroup);
      this.model.selectedGroup = null;
    }
  }

  addFqdnToHosts(): void {
    this.model.hosts.forEach((host) => {
      if (host.hostname && this.model.domain) {
        host.fqdn = `${host.hostname}.${this.model.domain}`;
      }
    });
  }

  onFinish(): void {
    this.addFqdnToHosts();
    console.log('Hosts with FQDN:', this.model.hosts);
    // Proceed with the final step...
  }

  // Called when the group is added from the group manager
  onGroupAdded() {
    if (this.showAddGroupForm) {
      this.groupManager.submitGroup().subscribe((data: any) => {
        if (data.id) {
          this.groups.push(data);
        }
      }, (data: any) => {
        if (data.status) {
          console.error('Error adding group:', data.status, data.message);
        } else {
          console.error('Error adding group:', data);
        }

      });
      this.showAddGroupForm = false;

    }
    this.wizard?.next();
  }
}


export interface Host {

  selectedIface: IfaceConfig | null; // Selected interface for the host
  iloIpAddr: string; // ILO IP address of the host
  iloPort: number; // ILO port of the host
  iloFqdn: string; // Fully Qualified Domain Name of the ILO
  username: string; // Username for the host
  password: string; // Password for the host
  selected?: boolean; // Optional property to track selection
  ifaceConfig?: IfaceConfig[]; // Optional property to store host configurations
  fqdn?: string; // Fully Qualified Domain Name of the host
}

export interface IfaceConfig {
  ifaceName: string; // Adapter name
  ipAddress: string; // IP address of the interface
  macAddress: string; // MAC address of the interface
  speed: string; // Speed of the interface (e.g., "10Gbits")
  status: string; // Status of the interface (e.g., "connected")
}
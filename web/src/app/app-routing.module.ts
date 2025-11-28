import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { DeploymentsComponent } from './deployments/deployments.component';
import { HealthChecksComponent } from './health-checks/health-checks.component';
import { SettingsComponent } from './settings/settings.component';
import { ManageUsersComponent } from './manage-users/manage-users.component';
import { HelpComponent } from './help/help.component';
import { ThemeComponent } from './settings/theme/theme.component';
import { LogsComponent } from './logs/logs.component';
import { LoginComponent } from './login/login.component';
import { ManageDhcpPoolsComponent } from './manage-dhcp-pools/manage-dhcp-pools.component';
import { ManageImagesComponent } from './manage-images/manage-images.component';
import { ManageGroupsComponent } from './manage-groups/manage-groups.component';
import { WizardComponent } from './wizard/wizard.component';
import { DhcpLeasesComponent } from './dhcp-leases/dhcp-leases.component';

const routes: Routes = [
  { path: 'deployments', component: DeploymentsComponent, children: [
    { path: '', redirectTo: 'wizard', pathMatch: 'full' }, // Default child route
    { path: 'manage-dhcp-pools', component: ManageDhcpPoolsComponent },
    { path: 'manage-images', component: ManageImagesComponent },
    { path: 'manage-groups', component: ManageGroupsComponent },
    { path: 'wizard', component: WizardComponent },
    { path: 'dhcp-leases', component: DhcpLeasesComponent }, // New route
  ]},
  { path: 'health-checks', component: HealthChecksComponent },
  { path: 'settings', component: SettingsComponent, children: [
    { path: '', redirectTo: 'help', pathMatch: 'full' }, // Default child route
    { path: 'help', component: HelpComponent },
    { path: 'manage-users', component: ManageUsersComponent },
    { path: 'logs', component: LogsComponent },
    { path: 'theme', component: ThemeComponent },
  ]},
  { path: 'login', component: LoginComponent },
  { path: '', redirectTo: '/deployments', pathMatch: 'full' }, // Default route
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }

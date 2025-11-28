import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { CommonModule } from '@angular/common';
import { ClarityModule } from '@clr/angular';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { CdsModule } from '@cds/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { HelpComponent } from './help/help.component';
import { ManageDhcpPoolsComponent } from './manage-dhcp-pools/manage-dhcp-pools.component';
import { ManageGroupsComponent } from './manage-groups/manage-groups.component';
import { ManageImagesComponent } from './manage-images/manage-images.component';
import { LogsComponent } from './logs/logs.component';
import { FlexLayoutModule } from "@angular/flex-layout";
import { ManageUsersComponent } from './manage-users/manage-users.component';
import { HostDeploymentComponent } from './wizard/host-deployment.component';
import { LoginComponent } from './login/login.component';
import { DeploymentsComponent } from './deployments/deployments.component';
import { HealthChecksComponent } from './health-checks/health-checks.component';
import { SettingsComponent } from './settings/settings.component';
import { DhcpLeasesComponent } from './dhcp-leases/dhcp-leases.component';
import { DhcpPoolManagerComponent } from './components/dhcp-pool-manager/dhcp-pool-manager.component';
import { ImageManagerComponent } from './components/image-manager/image-manager.component';
import { WizardComponent } from './wizard/wizard.component';
import { GroupManagerComponent } from './components/group-manager/group-manager.component';
import { ThemeComponent } from './settings/theme/theme.component';
@NgModule({
  declarations: [
    AppComponent,
    DeploymentsComponent,
    HealthChecksComponent,
    SettingsComponent,
    ManageUsersComponent,
    HelpComponent,
    LogsComponent,
    DhcpPoolManagerComponent,
    ManageDhcpPoolsComponent,
    ManageImagesComponent,
    ManageGroupsComponent,
    DhcpLeasesComponent,
    DhcpPoolManagerComponent,
    HostDeploymentComponent,
    ImageManagerComponent,
    WizardComponent,
  GroupManagerComponent,
  ThemeComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    ClarityModule,
    CdsModule,
    BrowserAnimationsModule,
    HttpClientModule,
    ReactiveFormsModule,
    FormsModule,
    FlexLayoutModule,
    CommonModule // Import CommonModule for directives like *ngFor
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }

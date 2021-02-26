import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ClarityModule } from "@clr/angular";
import { CdsModule } from '@cds/angular';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ManageIsoComponent } from './manage-iso/manage-iso.component';
import { ManageHostsComponent } from './manage-hosts/manage-hosts.component';
import { HelpComponent } from './help/help.component';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { ReactiveFormsModule } from '@angular/forms';
import { ManageDhcpComponent } from './manage-dhcp/manage-dhcp.component';

import '@cds/core/alert/register.js';
import '@cds/core/button/register.js';
import '@cds/core/accordion/register.js';


@NgModule({
  declarations: [
    AppComponent,
    ManageIsoComponent,
    ManageHostsComponent,
    HelpComponent,
    ManageDhcpComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
	ClarityModule,
	CdsModule,
	HttpClientModule,
	ReactiveFormsModule,
	BrowserAnimationsModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }

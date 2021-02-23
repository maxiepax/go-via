import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  constructor(private httpClient: HttpClient) { }

  public getHosts(){
  	return this.httpClient.get('http://172.16.214.15:8000/api/host');
  }

  public addHost(data){
  	return this.httpClient.post('http://172.16.214.15:8000/api/host', data);
  }  
  
  public getPools(){
  	return this.httpClient.get('http://172.16.214.15:8000/api/dhcp');
  }

  public addPool(data){
  	return this.httpClient.post('http://172.16.214.15:8000/api/dhcp', data);
  }
  
  public getImages(){
  	return this.httpClient.get('http://172.16.214.15:8000/api/image');
  }

  public addImage(data){
  	return this.httpClient.post('http://172.16.214.15:8000/api/image', data);
  }

}

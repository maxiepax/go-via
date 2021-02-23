import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  constructor(private httpClient: HttpClient) { }

  public getHosts(){
  	return this.httpClient.get('http://localhost:8080/api/host');
  }

  public addHost(data){
  	return this.httpClient.post('http://localhost:8080/api/host', data);
  }  
  
  public getPools(){
  	return this.httpClient.get('http://localhost:8080/v1/pools');
  }

  public addPool(data){
  	return this.httpClient.post('http://localhost:8080/v1/pools', data);
  }
  
  public getImages(){
  	return this.httpClient.get('http://localhost:8080/api/image');
  }

  public addImage(data){
  	return this.httpClient.post('http://localhost:8080/api/image', data);
  }

}

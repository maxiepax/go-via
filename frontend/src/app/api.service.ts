import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  constructor(private httpClient: HttpClient) { }

  public getGroups(){
  	return this.httpClient.get('http://localhost:8080/v1/groups');
  }

  public getHosts(){
  	return this.httpClient.get('http://localhost:8080/v1/addresses');
  }

  public addHost(data){
  	return this.httpClient.post('http://localhost:8080/v1/addresses', data);
  }  
  public deleteHost(id){
    return this.httpClient.delete(`http://localhost:8080/v1/addresses/${id}`);
  }

  public getPools(){
  	return this.httpClient.get('http://localhost:8080/v1/pools');
  }

  public addPool(data){
  	return this.httpClient.post('http://localhost:8080/v1/pools', data);
  }

  public deletePool(id){
    return this.httpClient.delete(`http://localhost:8080/v1/pools/${id}`);
  }
  
  public getImages(){
  	return this.httpClient.get('http://localhost:8080/api/image');
  }

  public addImage(data){
  	return this.httpClient.post('http://localhost:8080/api/image', data);
  }

  public getGroups(){
  	return this.httpClient.get('http://localhost:8080/v1/groups');
  }

  public addGroup(data){
  	return this.httpClient.post('http://localhost:8080/v1/groups', data);
  }

  public deleteGroup(id){
    return this.httpClient.delete(`http://localhost:8080/v1/groups/${id}`);
  }

}

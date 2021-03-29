import { Injectable } from '@angular/core';
import { HttpClient, HttpRequest, HttpEvent } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ApiService {
  constructor(private httpClient: HttpClient) {}

  public getHosts() {
    return this.httpClient.get(
      'http://' + window.location.hostname + ':8080/v1/addresses'
    );
  }

  public addHost(data) {
    return this.httpClient.post(
      'http://' + window.location.hostname + ':8080/v1/addresses',
      data
    );
  }

  public patchHost(id) {
    return this.httpClient.patch(
      'http://' + window.location.hostname + `:8080/v1/addresses/${id}`,
      {
        reserved: true
      }
    );
  }

  public deleteHost(id) {
    return this.httpClient.delete(
      'http://' + window.location.hostname + `:8080/v1/addresses/${id}`
    );
  }

  public getPools() {
    return this.httpClient.get(
      'http://' + window.location.hostname + ':8080/v1/pools'
    );
  }

  public addPool(data) {
    return this.httpClient.post(
      'http://' + window.location.hostname + ':8080/v1/pools',
      data
    );
  }

  public updatePool(id, data) {
    return this.httpClient.patch(
      `http://${window.location.hostname}:8080/v1/pools/${id}`,
      data
    );
  }

  public deletePool(id) {
    return this.httpClient.delete(
      'http://' + window.location.hostname + `:8080/v1/pools/${id}`
    );
  }

  public getGroups() {
    return this.httpClient.get(
      'http://' + window.location.hostname + ':8080/v1/groups'
    );
  }

  public addGroup(data) {
    return this.httpClient.post(
      'http://' + window.location.hostname + ':8080/v1/groups',
      data
    );
  }

  public deleteGroup(id) {
    return this.httpClient.delete(
      'http://' + window.location.hostname + `:8080/v1/groups/${id}`
    );
  }

  public getImages() {
    return this.httpClient.get(
      'http://' + window.location.hostname + ':8080/v1/images'
    );
  }

  public addImage(file: File): Observable<HttpEvent<any>> {
    const formData: FormData = new FormData();

    formData.append('file[]', file);

    const req = new HttpRequest(
      'POST',
      'http://' + window.location.hostname + `:8080/v1/images`,
      formData,
      {
        reportProgress: true,
        responseType: 'json'
      }
    );

    return this.httpClient.request(req);
  }

  public deleteImage(id) {
    return this.httpClient.delete(
      'http://' + window.location.hostname + `:8080/v1/images/${id}`
    );
  }
}

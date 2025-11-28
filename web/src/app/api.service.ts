

import { Injectable } from '@angular/core';
import { HttpClient, HttpRequest, HttpEvent } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { map } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class ApiService {

  constructor(private httpClient: HttpClient) {}

  public getThemeImage(): Observable<Blob> {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/theme/image',
      { responseType: 'blob' }
    );
  }


  public getHosts() {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/hosts'
    );
  }

  public getHost(id) {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/hosts/' + id
    );
  }

  public addHost(data) {
    return this.httpClient.post(
      'https://' + window.location.host + '/v1/hosts',
      data
    );
  }

  public updateHost(id, data) {
    return this.httpClient.patch(
      `https://${window.location.host}/v1/hosts/${id}`,
      data
    );
  }



  public reimageHost(id) {
    return this.httpClient.patch(
      'https://' + window.location.host + `/v1/hosts/${id}`,
      {
        reimage: true,
        progress: 0,
        progresstext: 'reimaging',
      }
    );
  }

  public cancelImageHost(id) {
    return this.httpClient.patch(
      'https://' + window.location.host + `/v1/hosts/${id}`,
      {
        reimage: false,
        progress: 0,
        progresstext: 'reimaging canceled',
      }
    );
  }

  uploadBackgroundImage(formData: FormData) {
      return this.httpClient.post(
        'https://' + window.location.host + '/v1/theme/image',
        formData
      );
    }


  public deleteHost(id) {
    return this.httpClient.delete(
      'https://' + window.location.host + `/v1/hosts/${id}`
    );
  }

  public getPools() {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/pools'
    );
  }

  public addPool(data) {
    return this.httpClient.post(
      'https://' + window.location.host + '/v1/pools',
      data
    );
  }

  public updatePool(id, data) {
    return this.httpClient.patch(
      `https://${window.location.host}/v1/pools/${id}`,
      data
    );
  }

  public deletePool(id) {
    return this.httpClient.delete(
      'https://' + window.location.host + `/v1/pools/${id}`
    );
  }

  public getGroups() {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/groups'
    );
  }

  public addGroup(data) {
    return this.httpClient.post(
      'https://' + window.location.host + '/v1/groups',
      data
    );
  }

  public updateGroup(id, data) {
    return this.httpClient.patch(
      `https://${window.location.host}/v1/groups/${id}`,
      data
    );
  }

  public deleteGroup(id) {
    return this.httpClient.delete(
      'https://' + window.location.host + `/v1/groups/${id}`
    );
  }

  public getImages() {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/images'
    );
  }

  public addImage(file: File, hash: string, description: string): Observable<HttpEvent<any>> {
    const formData: FormData = new FormData();

    formData.append('file[]', file);
    formData.append('hash', hash || "");
    formData.append('description', description || "");

    const req = new HttpRequest(
      'POST',
      'https://' + window.location.host + `/v1/images`,
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
      'https://' + window.location.host + `/v1/images/${id}`
    );
  }

  public getUsers() {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/users'
    );
  }

  public addUser(data) {
    return this.httpClient.post(
      'https://' + window.location.host + '/v1/users',
      data
    );
  }

  public updateUser(id, data) {
    return this.httpClient.patch(
      `https://${window.location.host}/v1/users/${id}`,
      data
    );
  }

  public deleteUser(id) {
    return this.httpClient.delete(
      'https://' + window.location.host + `/v1/users/${id}`
    );
  }

  public getVersion() {
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/version'
    );
  }

  public checkILOM(ip, port) {
    return this.httpClient.post(
      'https://' + window.location.host + '/v1/ilohosts/checkilo',
      {
        iloIpAddr: ip,
        port: String(port),
      }
    );
  }

  public restartHost(hostId, payload) {
    return this.httpClient.post(
      `https://${window.location.host}/v1/ilohosts/${hostId}/reboot`,
      payload
    );
  }
  public addHostVlan(hostId, payload)  {
    console.log(payload)
    return this.httpClient.post(
      `https://${window.location.host}/v1/ilohosts/${hostId}/setvlanID`,
      payload
    );
  }
  public shutdownHost(hostId, payload) {
    return this.httpClient.post(
      `https://${window.location.host}/v1/ilohosts/${hostId}/shutdown`,
      payload
    );
  }

  public getServerPowerState(hostId, payload): Observable<any> {
    if (!hostId) {
      console.error('Invalid host ID for getting power state:', hostId);
      return of({ error: 'Invalid host ID' });
    }
    return this.httpClient.post(
      `https://${window.location.host}/v1/ilohosts/${hostId}/powerstate`,
      payload
    );
  }
  public startHost(hostId, payload) {
    console.log(payload)
    return this.httpClient.post(
      `https://${window.location.host}/v1/ilohosts/${hostId}/start`,
      payload
    );
  }

  public setOneTimeBoot(hostId, payload) {
    return this.httpClient.post(
      `https://${window.location.host}/v1/hosts/${hostId}/onetimeboot`,
      payload
    );
  }

  public getHostConfig(ip: string, port: number, apiFlavour: string, username: string, password: string): Observable<any> {
    // Log the parameters to the console
    console.log(`getHostConfig called with ip: ${ip}, port: ${port}, apiFlavour: ${apiFlavour}`);

    // Return a mocked successful response
    return this.httpClient.get(
      'https://' + window.location.host + '/v1/hostconfig',
      {
        params: {
          iloIpAddr: ip,
          port: String(port),
          apiFlavour: apiFlavour,
          username: username,
          password: password,
        }
      }
    );
  }

  public getFormattedAddresses(): Observable<any[]> {
    return this.httpClient.get<any[]>(
      `https://${window.location.host}/v1/hosts`
    ).pipe(
      map((addresses) => {
        return addresses.map(address => ({
          ip: address.ip,
          hostname: address.hostname,
          domain: address.domain,
          expires_at: address.expires_at,
        }));
      })
    );
  }
}

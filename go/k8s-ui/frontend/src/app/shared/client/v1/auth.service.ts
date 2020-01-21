import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {HttpClient, HttpHeaders} from '@angular/common/http';

@Injectable()
export class AuthoriseService {
  headers = new HttpHeaders({'Content-type': 'application/json'});
  options = {headers: this.headers};

  constructor(private http: HttpClient) {}

  login(username: string, password: string, type: string): Observable<any> {
    const url = `login/${type}?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`;
    return this.http.post(url, null, this.options);
  }
}

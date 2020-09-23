import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {HttpClient, HttpHeaders} from '@angular/common/http';

@Injectable()
export class AuthoriseService {
  headers = new HttpHeaders({'Content-Type': 'application/json'});
  options = {headers: this.headers};

  constructor(private http: HttpClient) {}

  login(username: string, password: string, type: string): Observable<any> {
    // const url = `login/${type}?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`;
    const url = `login/${type}`;
    return this.http.post(url, {
      username,
      password,
    }, this.options);
  }
}

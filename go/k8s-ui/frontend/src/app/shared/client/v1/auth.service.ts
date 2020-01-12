import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {HttpClient, HttpHeaders} from '@angular/common/http';

@Injectable()
export class AuthoriseService {
  headers = new HttpHeaders({'Content-type': 'application/json'});
  options = {'headers': this.headers};
  
  constructor(private http: HttpClient) {}
  
  login(username: string, password: string, type: string): Observable<any> {
    return this.http.post(`login/${type}?username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`, null, this.options);
  }
}

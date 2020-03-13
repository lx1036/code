import {Injectable} from '@angular/core';
import {Observable} from "rxjs";
import {CsrfToken} from "../../../typings/backend-api";
import {HttpClient} from "@angular/common/http";

@Injectable()
export class CsrfTokenService {

  constructor(private readonly http: HttpClient) {}

  /** Get a CSRF token for an action you want to perform. */
  getTokenForAction(action: string): Observable<CsrfToken> {
    return this.http.get<CsrfToken>(`api/v1/csrftoken/${action}`);
  }
}

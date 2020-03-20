import {Injectable} from "@angular/core";
import {AuthResponse, CsrfToken, K8SError, LoginSpec} from "../../../typings/backend-api";
import {Observable, of} from "rxjs";
import {CsrfTokenService} from "./csrftoken";
import {switchMap} from "rxjs/operators";
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {CONFIG} from "../../../index.config";
import {CookieService} from 'ngx-cookie-service';

@Injectable()
export class AuthService {
  private readonly _config = CONFIG;

  constructor(
    private readonly csrfTokenService: CsrfTokenService,
    private readonly http: HttpClient,
    private readonly cookies_: CookieService,) {
  }

  /**
   * Sends a login request to the backend with filled in login spec structure.
   */
  login(loginSpec: LoginSpec): Observable<K8SError[]> {
    return this.csrfTokenService
    .getTokenForAction('login')
    .pipe(
      switchMap((csrfToken: CsrfToken) =>
        this.http.post<AuthResponse>('api/v1/login', loginSpec, {
          headers: new HttpHeaders().set(this._config.csrfHeaderName, csrfToken.token),
        }),
      ),
    )
    .pipe(
      switchMap((authResponse: AuthResponse) => {
        if (authResponse.jweToken.length !== 0 && authResponse.errors.length === 0) {
          this.setTokenCookie_(authResponse.jweToken);
        }

        return of(authResponse.errors);
      }),
    );
  }

  private setTokenCookie_(token: string): void {
    // This will only work for HTTPS connection
    this.cookies_.set(this._config.authTokenCookieName, token, null, null, null, true, 'Strict');
    // This will only work when accessing Dashboard at 'localhost' or
    // '127.0.0.1'
    this.cookies_.set(
      this._config.authTokenCookieName,
      token,
      null,
      null,
      'localhost',
      false,
      'Strict',
    );
    this.cookies_.set(
      this._config.authTokenCookieName,
      token,
      null,
      null,
      '127.0.0.1',
      false,
      'Strict',
    );
  }
}




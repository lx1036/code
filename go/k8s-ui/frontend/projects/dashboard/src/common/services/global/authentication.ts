import {Injectable} from "@angular/core";
import {AuthResponse, CsrfToken, K8SError, LoginSpec} from "../../../typings/backend-api";
import {Observable, of} from "rxjs";
import {CsrfTokenService} from "./csrftoken";
import {switchMap} from "rxjs/operators";
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {CONFIG} from "../../../index.config";


@Injectable()
export class AuthService {
  private readonly _config = CONFIG;

  constructor(
    private readonly csrfTokenService: CsrfTokenService,
    private readonly http: HttpClient,) {
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
}




import {
  HttpClient,
  HttpErrorResponse,
  HttpEvent,
  HttpHandler, HttpHeaders,
  HttpInterceptor,
  HttpRequest
} from "@angular/common/http";
import {TranslateHttpLoader} from "@ngx-translate/http-loader";
import {AuthService} from "./shared/components/auth/auth.service";
import {Injectable, Injector} from "@angular/core";
import {httpStatusCode} from "./shared/shared.const";
import {Router} from "@angular/router";
import {Location} from "@angular/common";
import {Observable} from "rxjs";
import {LoginTokenKey} from "./shared/components/auth/auth.const";
import {environment} from "../environments/environment";


export function HttpLoaderFactory(httpClient: HttpClient) {
  return new TranslateHttpLoader(httpClient, './assets/i18n/', '.json');
}

export function initUser(authService: AuthService, injector: Injector) {
  return () => authService.retrieveUser().catch((error: HttpErrorResponse) => {
    console.log('error status: ' + error.status);
    if (error.status ===  httpStatusCode.Unauthorized) {
      injector.get(Router).navigateByUrl('sign-in')
        .then(() => console.log('navigate to sign-in page'))
        .catch((err: HttpErrorResponse) => console.log("can't navigate to sign-in page because of:"  + err.message ));
    }
  });
}

export function initConfig(authService: AuthService) {
  return () => authService.initConfig();
}

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private location: Location) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const token = localStorage.getItem(LoginTokenKey);
    const headers: {[name: string]: string|string[]} = {};
    for (const key of request.headers.keys()) {
      headers[key] = request.headers.getAll(key);
    }

    // headers['Content-Type'] = 'application/json';

    if (token) { // if logged in
      headers.Authorization = 'Bearer ' + token;
    }

    let url = request.url;
    if (request.url.indexOf('assets') === -1) {
      url = this.location.normalize(environment.api) + this.location.prepareExternalUrl(request.url);
    }

    const req = request.clone({url, headers: new HttpHeaders(headers)});

    return next.handle(req).pipe();
  }
}

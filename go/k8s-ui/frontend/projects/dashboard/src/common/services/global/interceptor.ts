

import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {CookieService} from 'ngx-cookie-service';
import {Observable} from 'rxjs/Observable';
import {CONFIG} from '../../../index.config';

/* tslint:disable */
// We can disable tslint for this file as any is required here.
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private readonly cookies_: CookieService) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const authCookie = this.cookies_.get(CONFIG.authTokenCookieName);
    // Filter requests made to our backend starting with 'api/v1' and append request header
    // with token stored in a cookie.
    if (req.url.startsWith('api/v1') && authCookie.length) {
      const authReq = req.clone({
        headers: req.headers.set(CONFIG.authTokenHeaderName, authCookie),
      });

      return next.handle(authReq);
    }

    return next.handle(req);
  }
}
/* tslint:enable */

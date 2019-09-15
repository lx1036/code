import {Injectable, Injector} from '@angular/core';
import {HttpClient, HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '../packages/angular/common/http';
import {Observable, throwError} from 'rxjs';
import {User} from '../models/user';
import {CanActivate, Router} from '@angular/router';
import {catchError} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private BASE_URL = 'http://localhost:1337';
  constructor(private _http: HttpClient) { }

  getToken(): string {
    return localStorage.getItem('token');
  }

  logIn(email: string, password: string): Observable<any> {
    const url = `${this.BASE_URL}/login`;

    return this._http.post<User>(url, {email, password});
  }

  signUp(email: string, password: string): Observable<User> {
    const url = `${this.BASE_URL}/register`;

    return this._http.post<User>(url, {email, password});
  }

  getStatus(): Observable<User> {
    const url = `${this.BASE_URL}/status`;
    console.log(url);

    return this._http.get<User>(url);
  }
}


@Injectable({
  providedIn: 'root'
})
export class TokenInterceptor implements HttpInterceptor {
  private authService: AuthService;
  constructor(private injector: Injector) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    this.authService = this.injector.get(AuthService);
    const token: string = this.authService.getToken();
    request = request.clone({
      setHeaders: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    return next.handle(request);
  }
}

@Injectable()
export class ErrorInterceptor implements HttpInterceptor {
  constructor(private _router: Router) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(request).pipe(
      catchError((response: any) => {
        if (response instanceof HttpErrorResponse && response.status === 401) {
          localStorage.removeItem('token');
          this._router.navigateByUrl('/log-in');
        }

        return throwError(response);
      })
    );
  }
}

@Injectable({
  providedIn: 'root'
})
export class AuthGuardService implements CanActivate {
  constructor(public auth: AuthService, public router: Router) {}

  canActivate(): boolean {
    if (!this.auth.getToken()) {
      this.router.navigateByUrl('/log-in');
      return false;
    }

    return true;
  }
}
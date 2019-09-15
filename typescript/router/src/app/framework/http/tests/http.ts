import {Component, Injectable, NgModule, OnInit} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HTTP_INTERCEPTORS, HttpClient, HttpClientModule, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Observable} from 'rxjs';


/**
 *
 * HttpClient
 *
 * HttpInterceptor
 * HttpHandlerInterface(HttpBackend): httpResponse: HttpResponse = httpHandler(request: HttpRequest, httpInterceptor)
 *
 */






@Injectable({
  providedIn: 'root'
})
export class AInterceptor implements HttpInterceptor {
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const request = req.clone({setHeaders: {'User-ID': '1'}});

    return next.handle(request);
  }

}


@Injectable({
  providedIn: 'root'
})
export class BInterceptor implements HttpInterceptor{
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const request = req.clone({setHeaders: {'User-ID-Next': '2'}});

    return next.handle(request);
  }
}




@Component({
  selector: 'demo-http',
  template: `
    <div>
      <pre>
        {{posts | json}}
      </pre>
    </div>
  `
})
export class DemoHttp implements OnInit {
  public posts;

  public constructor(private http: HttpClient) {
  }

  ngOnInit() {
    this.http.get('https://jsonplaceholder.typicode.com/posts/1', {observe: 'events'}).subscribe(res => {
      this.posts = res;
    });
  }
}


@NgModule({
  declarations: [
    DemoHttp,
  ],
  imports: [
    BrowserModule,
    HttpClientModule
  ],
  providers: [
    {provide: HTTP_INTERCEPTORS, useClass: AInterceptor, multi: true},
    {provide: HTTP_INTERCEPTORS, useClass: BInterceptor, multi: true},
  ],
  bootstrap: [DemoHttp]
})
export class DemoHttpModule {

}
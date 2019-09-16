import {Component, NgModule, OnInit} from "@angular/core";
import {BrowserModule} from "@angular/platform-browser";
import {HttpClient, CustomHttpModule} from "../src/handler";
import {Request, Response} from "../src/model";
import {HTTP_INTERCEPTORS, HttpHandler, Interceptor} from "../src/handler";




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
    this.http.get('https://jsonplaceholder.typicode.com/posts/1').subscribe((res: Response<any>) => {
      console.log(res);
      this.posts = res;
    });
  }
}


export class CustomInterceptor implements Interceptor{
  intercept(request: Request<any>, next: HttpHandler) {
    request = request.clone({
      setHeaders: {'Custom-Header-1': ['a']}
    });

    return next.handle(request);
  }
}



@NgModule({
  declarations: [
    DemoHttp,
  ],
  imports: [
    BrowserModule,

    CustomHttpModule,
  ],
  providers: [
    {provide: HTTP_INTERCEPTORS, useClass: CustomInterceptor, multi: true}
  ],
  bootstrap: [DemoHttp]
})
export class TestCustomHttpClientModule {

}



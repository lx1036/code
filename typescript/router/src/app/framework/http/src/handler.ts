import {Observable, Observer, of, Subscriber} from "rxjs";
import {Injectable, InjectionToken, Injector, NgModule} from "@angular/core";
import {Request, HttpEvent, ErrorResponse, Response, Headers, EventType, HeaderResponse, Methods} from "./model";
import {concatMap} from "rxjs/operators";




export abstract class HttpHandler {
  abstract handle(request: Request<any>): Observable<HttpEvent<any>>;
}


export abstract class HttpBackend extends HttpHandler{

}

export const HTTP_INTERCEPTORS = new InjectionToken<Interceptor[]>('HTTP_INTERCEPTORS');


export abstract class Interceptor {
  abstract intercept(request: Request<any>, next: HttpHandler);
}

export class NoopInterceptor implements Interceptor {
  intercept(request: Request<any>, next: HttpHandler) {
    return next.handle(request);
  }
}


/**
 * TODO:
 * Subscriber implements Observer
 * Subscription
 */

/**
 * This instance is mainly sending real request to backend
 */
@Injectable()
export class XhrBackend implements HttpBackend {
  handle(request: Request<any>): Observable<HttpEvent<any>> {

    return new Observable<HttpEvent<any>>((subscriber: Subscriber<HttpEvent<any>>) => {
      /**
       * xhr spec: https://developer.mozilla.org/zh-CN/docs/Web/API/XMLHttpRequest
       * event types: onload, onloadstart, onloadend, onerror, onprogress, onabort
       * using: https://developer.mozilla.org/zh-CN/docs/Web/API/XMLHttpRequest/Using_XMLHttpRequest
       */
      const xhr = new XMLHttpRequest();
      xhr.open(request.method, request.url, true);

      if (!! request.withCredentials) {
        xhr.withCredentials = true;
      }

      request.headers.forEach((name, values: string[]) => xhr.setRequestHeader(name, values.join(',')));


      // request.headers.forEach((value, header) => xhr.setRequestHeader(header, value));

      const getHeaderResponse = () => {
        /**
         * pragma: no-cache
         * content-type: application/json; charset=utf-8
         * cache-control: public, max-age=14400
         * expires: Sat, 08 Sep 2018 17:04:37 GMT
         */
        const headers = new Headers(xhr.getAllResponseHeaders());
        const status = xhr.status;
        const statusText = xhr.statusText;
        const url = xhr.responseURL || request.url;

        return new HeaderResponse({headers, status, statusText, url});
      };

      /**
       * The 'onload' event handler, meaning the response is available
       */
      const onLoad = (event: Event) => {
        // const {headers, status, statusText, url} = getHeaderResponse();

        /**
         * TODO: remove this
         */
        const headers = new Headers(xhr.getAllResponseHeaders());
        const status = xhr.status;
        const statusText = xhr.statusText;
        const url = xhr.responseURL || request.url;

        subscriber.next(new Response(xhr.response, headers, xhr.status, xhr.statusText));
        subscriber.complete();
      };

      const onError = (event: ErrorEvent) => {
        subscriber.error(new ErrorResponse({error: event, status: xhr.status, statusText: xhr.statusText || 'Unknown Error'}));
      };

      /**
       * The download progress event handler.
       */
      const onDownloadProgress = (event: ProgressEvent) => {};
      /**
       * The upload progress event handler.
       */
      const onUploadProgress = (event: ProgressEvent) => {};
      xhr.addEventListener('load', onLoad);
      xhr.addEventListener('error', onError);

      const requestBody = request.serializeBody();

      if (request.isProgressed) {
        xhr.addEventListener('progress', onDownloadProgress);

        if (xhr.upload && requestBody !== null) {
          xhr.upload.addEventListener('progress', onUploadProgress);
        }
      }

      // send the request, and notify the event stream
      // xhr.send(requestBody);
      // subscriber.next({type: EventType.Sent});

      return () => {};
    });
  }
}




/**
 * Decorator Pattern
 *
 * Tree Structure
 */
export class HttpInterceptorHandler implements HttpHandler {
  constructor(private _next: HttpHandler, private _interceptor: Interceptor) {}

  handle(request: Request<any>): Observable<HttpEvent<any>> {
    return this._interceptor.intercept(request, this._next);
  }
}


@Injectable()
export class HttpInterceptingHandler implements HttpHandler {
  constructor(private _injector: Injector, private _backend: HttpBackend) {}

  handle(request: Request<any>): Observable<HttpEvent<any>> {
    const interceptors = this._injector.get(HTTP_INTERCEPTORS);
    /**
     * [interceptor1, interceptor2, interceptor3] ->
     * Register:
     * handler = XhrBackend ->
     * interceptor_handler1 = InterceptorHandler(handler, interceptor3) ->
     * interceptor_handler2 = InterceptorHandler(interceptor_handler1, interceptor2) ->
     * interceptor_handler3 = InterceptorHandler(interceptor_handler2, interceptor1) ->
     * interceptor_handler3.handle(request)
     *
     *
     * Register: handler = new InterceptorHandler(new InterceptorHandler(new InterceptorHandler(xhrHandler, interceptor3), interceptor2), interceptor1)
     * Run: handler.handle(request)
     * interceptor1.intercept(request, new InterceptorHandler(new InterceptorHandler(xhrHandler, interceptor3), interceptor2)) =
     * new InterceptorHandler(new InterceptorHandler(xhrHandler, interceptor3), interceptor2).handle(request)
     * interceptor2.intercept(request, new InterceptorHandler(xhrHandler, interceptor3)) = new InterceptorHandler(xhrHandler, interceptor3).handle(request)
     * interceptor3.intercept(request, xhrHandler) = xhrHandler.handle(request): Observable<HttpEvent<any>>
     *
     *
     * Run:
     * interceptor_handler3.handle(request) -> interceptor_handler3.intercept(request, interceptor_handler2)
     * -> interceptor_handler2.handle(request) -> interceptor_handler1.intercept(request, interceptor_handler)
     * ->interceptor_handler.handle(request) = XhrBackend.handle(request)
     */
    const chain: HttpHandler = interceptors.reduceRight(
      (handler: HttpHandler, interceptor: Interceptor) => new HttpInterceptorHandler(handler, interceptor),
      this._backend
    );

    return chain.handle(request);
  }
}






/**
 * *******************************************MODULE********************************************************************
 */

/**
 * Perform HTTP requests.
 *
 */
@Injectable()
export class HttpClient {
  constructor(private handler: HttpHandler) {}
  
  request(method: Methods, uri: string, options?: {}): Observable<HttpEvent<any>> {
    let request = new Request(method, uri, options);
    
    let response = of(request).pipe(concatMap(
      (request: Request<any>): Observable<HttpEvent<any>> => this.handler.handle(request)
    ));
    
    return response;
  }
  
  get(uri: string): Observable<any> {
    return this.request('GET', uri);
  }
}

/**
 * Handler(send xhr request) + Interceptor(intercept xhr request, do something)
 */
@NgModule({
  providers: [
    HttpClient,
    {provide: HttpHandler, useClass: HttpInterceptingHandler},
    {provide: HttpBackend, useClass: XhrBackend},
    {provide: HTTP_INTERCEPTORS, useClass: NoopInterceptor, multi: true}
  ]
})
export class CustomHttpModule {

}


export class Headers {
  headers: Map<string, string[]>;


  constructor(headers?: string) {
    this.headers = new Map<string, string[]>();

    if (!! headers) {
      headers.split('\n').forEach(line => {
        const index = line.indexOf(':');
        const name = line.slice(0, index);
        const value = line.slice(index + 1).trim();

        this.headers.set(name, [value]);
      });
    }
  }

  forEach(callback: (name: string, values: string[]) => void) {
      this.headers.forEach((values, name) => {

        callback(name, values);
      });
  }

  set(name: string, values: string[]) {
    this.headers.set(name, values);
  }
}


export class Params {

}

/**
 * Outgoing HTTP request
 */
export class Request<T> {
  withCredentials: boolean;
  headers: Headers;

  method: string;


  readonly isProgressed = false;

  // constructor(method: 'POST'|'PUT'|'PATCH', url: string, body: T|null, init?:{
  //   headers?: Headers,
  //   params?: Params,
  // });
  constructor(method: string, readonly url: string, init?: {
    headers?: Headers,
  }) {
    this.method = method.toUpperCase();

    if (init) {
      if (!! init.headers) {
        this.headers = init.headers;
      }
    }

    if (! this.headers) {
      this.headers = new Headers();
    }
  }

  clone(options: {
    method?: string,
    headers?: Headers,
    setHeaders?: {[name: string]: string[]},
  } = {}): Request<T> {
    const method = options.method || this.method;
    let headers = options.headers || this.headers;

    if (!! options.setHeaders) {
      Object.keys(options.setHeaders).forEach((name) => {
        headers.set(name, options.setHeaders[name]);
      });
    }

    return new Request<T>(method, this.url, {headers});
  }

  serializeBody() {

  }
}



export type Methods = 'GET'|'POST'|'PUT'|'DELETE'|'HEAD'|'OPTIONS';



export enum EventType {
  Sent,
}

interface HttpSentEvent {
  type: EventType.Sent
}

export type HttpEvent<T> = HttpSentEvent | Response<T>;




export class Response<T> {
  constructor(private body, private headers: Headers, private status, private statusText) {}
}

export class ErrorResponse {
  constructor(init?: {error?: any, status?: number, statusText?: string}) {}
}

export class HeaderResponse {
  constructor(public init: {headers: Headers, status: number, statusText: string, url: string}) {

  }
}
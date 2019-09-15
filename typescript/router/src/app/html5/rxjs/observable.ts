import {Observable, of, Subscriber, TeardownLogic} from "rxjs";
import {concatMap, take} from "rxjs/operators";


const httpEvents$ = new Observable((observer) => {
  observer.next('b');

  return () => { return 'a'};
});

httpEvents$.subscribe((event) => {
  console.log(event);
});


/**
 *
 * *******************************************of()/concatMap()**********************************************************
 */

function request(...options: string[]): Observable<string[]> {
  const response: Observable<string[]> = of(options).pipe(
    concatMap((options: string[]) =>
      new Observable<string[]>((subscriber: Subscriber<string[]>): TeardownLogic => {
        subscriber.next(options);
  
        return (): void => {}
      })
    ),
    take(2),
);
  
  return response;
}

request('a', 'b', 'c').subscribe(value => console.log(value));
request('d', 'e', 'f').subscribe(value => console.log(value));



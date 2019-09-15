function Observable(subscribe) {
  this.subscribe = subscribe;
}
const one$ = new Observable(observer => {
  observer.next(1);
  observer.complete();
});

one$.subscribe({
  next: value => console.log(value), // 1
  complete: () => {}
});
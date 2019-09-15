/**
 * Data structure over time
 */

import {Observable} from 'rxjs';
import {map} from "rxjs/operators";

export class MyBehaviorSubject extends Observable {
  observers = [];
  lastValue;
  
  constructor(initialValue) {
    super();
    
    if (typeof initialValue === 'undefined') {
      throw new Error('You need to provide initial value');
    }
    
    this.lastValue = initialValue;
  }
  
  // subscribe(observer) {
  //   this.observers.push(observer);
  //   observer.next(this.lastValue);
  // }
  
  next(value) {
    this.lastValue = value;
    this.observers.forEach(observer => observer.next(value));
  }
  
  getValue() {
    return this.lastValue;
  }
}

const subject = new MyBehaviorSubject('initialValue');

subject.pipe(map(value => `Observer one ${value}`)).subscribe(function(value) {
  console.log(value);
});

subject.next('New value');

// setTimeout(() => {
//   subject.map(value => `Observer two ${value}`).subscribe(function(value) {
//     console.log(value);
//   });
// }, 2000);
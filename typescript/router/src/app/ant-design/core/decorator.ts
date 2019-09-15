import {coerceBooleanProperty} from '@angular/cdk/coercion';


export function propDecoratorFactory(name: string, callback) {
  const propDecorator = (target: any, propName: string) => {
    const privatePropName = `$$__${propName}`;

    Object.defineProperty(target, propName, {
      set(value: any): void {
        this[privatePropName] = callback(value);
      },
      get(): any {
        return this[privatePropName];
      }
    });
  };

  return propDecorator;
}

/*const callback = (value) => {
  return value + '/b';
};*/


export function InputBoolean() {
  return propDecoratorFactory('InputBoolean', coerceBooleanProperty);
}


/*export class TestClass {
  @InputBoolean() static a = 'a';
  @InputBoolean() b = 'a';
}

console.log(TestClass.a, (new TestClass()).b);*/




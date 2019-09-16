
import 'reflect-metadata';

/**
 * @see https://www.tslang.cn/docs/handbook/decorators.html
 * @see https://github.com/rbuckton/reflect-metadata
 * @see http://es6.ruanyifeng.com/#docs/decorator
 */



@Reflect.metadata('component', 'HelloComp')
class C {
  @Reflect.metadata('a', 1)
  method() {
  }
  
  componentRef;
}
Reflect.defineMetadata('b', 2, C.prototype, 'componentRef');
let obj = new C();
console.log(Reflect.getMetadata('a', obj, 'method'),
  Reflect.getMetadata('b', obj, 'componentRef'),
  Reflect.getMetadata('component', C),
  Reflect.getMetadataKeys(C));


/**
 * @link http://es6.ruanyifeng.com/#docs/decorator
 */
@testable
class MyTestableClass {
  // ...
}
function testable(target) {
  target.isTestable = true;
}
console.log(MyTestableClass.isTestable); // true


@decorator
class A {}
function decorator(target) {
  target.isTestable = true;
}
// Equals to
let B = decorator(A) || A;
console.log(B.isTestable);


function testable2(isTestable) {
  return function(target) {
    target.isTestable = isTestable;
  }
}
@testable2(true)
class MyTestableClass2 {}
console.log(MyTestableClass.isTestable); // true
@testable2(false)
class MyClass {}
console.log(MyClass.isTestable); // false


function testable3(target) {
  target.prototype.isTestable = true;
}
@testable3
class MyTestableClass3 {}
let obj2 = new MyTestableClass3();
console.log(obj2.isTestable); // true


export function mixins(...list) {
  return function (target) {
    Object.assign(target.prototype, ...list)
  }
}
const Foo = {
  foo() { console.log('foo') }
};
@mixins(Foo)
class MyClass2 {}
let obj3 = new MyClass2();
obj3.foo(); // 'foo'
Object.assign(MyClass2.prototype, Foo);
obj3.foo();


function dec(id){
  console.log('evaluated', id);
  return (target, property, descriptor) => console.log('executed', id);
}
class Example {
  @dec(1)
  @dec(2)
  method(){}
}


// Decorator can't be used in Function
var counter = 0;
function add () {
  counter++;
}
@add()
function foo() {
}
foo();
console.log(counter);


export interface Directive {

}

export interface Directive {

}
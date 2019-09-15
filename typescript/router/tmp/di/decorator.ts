

/**
 * yarn ts-node src/app/di/decorator.ts
 */


/*
export function makeDecorator(name: string, props?: (...args: any[]) => any, parentClass?: any, chainFn?: (fn: Function) => void, typeFn?: (type: Type<any>, ...args: any[]) => void):
  {new (...args: any[]): any; (...args: any[]): any; (...args: any[]): (cls: any) => any;} {
  const metaCtor = makeMetadataCtor(props);

  function DecoratorFactory(...args: any[]): (cls: any) => any {
    console.log(args);

    if (this instanceof DecoratorFactory) {
      metaCtor.call(this, ...args);
      return this;
    }

    const annotationInstance = new (<any>DecoratorFactory)(...args);
    const TypeDecorator: TypeDecorator = <TypeDecorator>function TypeDecorator(cls: Type<any>) {
      typeFn && typeFn(cls, ...args);
      // Use of Object.defineProperty is important since it creates non-enumerable property which
      // prevents the property is copied during subclassing.
      const annotations = cls.hasOwnProperty(ANNOTATIONS) ? (cls as any)[ANNOTATIONS] : Object.defineProperty(cls, ANNOTATIONS, {value: []})[ANNOTATIONS];
      annotations.push(annotationInstance);

      return cls;
    };

    if (chainFn) chainFn(TypeDecorator);

    return TypeDecorator;
  }

  if (parentClass) {
    DecoratorFactory.prototype = Object.create(parentClass.prototype);
  }

  DecoratorFactory.prototype.ngMetadataName = name;

  // console.log(DecoratorFactory, typeof DecoratorFactory);

  (<any>DecoratorFactory).annotationCls = DecoratorFactory;

  return DecoratorFactory as any;
}

function makeMetadataCtor(props?: (...args: any[]) => any): any {
  return function ctor(...args: any[]) {
    if (props) {
      const values = props(...args);
      for (const propName in values) {
        this[propName] = values[propName];
      }
    }
  };
}


export const Directive: DirectiveDecorator = makeDecorator('Directive', (dir: Directive = {}) => dir);



*/
// console.log(Directive, Directive({selector: 'test'}), new AppDirective());


/**
 *
 */

export const ANNOTATIONS = '__annotations__';
export const PARAMETERS = '__parameters__';
export const PROP_METADATA = '__prop__metadata__';


function makeMetadata(props?: (...args: any[]) => any): (...args: any[]) => any {
  return (...args: any[]) => {
    if (props) {
      const value = props(...args);
      
      
    }
  };
}

function decoratorFactory(...args: any[]): (...args: any[]) => any {
  
  if (this instanceof decoratorFactory) {
    return this;
  }
  
  console.log(args);
  
  const annotationInstance = new (decoratorFactory as any)(...args);
  
  return (className) => {
    console.log(className);
    
    const annotations = className.hasOwnProperty(ANNOTATIONS) ? className[ANNOTATIONS] :
      Object.defineProperty(className, ANNOTATIONS, {value: []})[ANNOTATIONS];
    
    console.log(annotations);
    annotations.push(annotationInstance);
    
    return className;
  };
}

function makeDecorator(name: string, props?: (...args: any[]) => any): (...args: any[]) => any {
  // const metadata = makeMetadata(props);
  
  console.log(name);
  
  decoratorFactory.prototype.ngMetadataName = name;
  (decoratorFactory as any).annotationClassName = decoratorFactory;
  
  return decoratorFactory;
}


interface DirectiveDecorator {
  new (obj: Directive): Directive;
}

export interface Directive {
  selector?: string;
}

export interface Component extends Directive {
}


export enum ChangeDetectionStrategy {
  OnPush = 0,
  Default = 1,
}
export const Directive = makeDecorator(
  'Directive', (dir: Directive = {}) => dir);

export const Component = makeDecorator(
  'Component', (c: Component = {}) => ({changeDetection: ChangeDetectionStrategy.Default, ...c}));



@Directive({
  selector: 'app-directive'
})
export class AppDirective {

}

@Component({
  selector: 'app-component'
})
export class AppComponent {

}





import {Type} from '../packages/angular/core/src/type';
import {ANNOTATIONS, TypeDecorator} from '../packages/angular/core/src/util/decorators';
// import {defineInjector, InjectorType, ModuleWithProviders, Provider, SchemaMetadata} from '@angular/core';
import {convertInjectableProviderToFactory} from '../packages/angular/core/src/di/injectable';
import {Provider} from '../packages/angular/core/src/di/provider';
import {ModuleWithProviders, SchemaMetadata} from '../packages/angular/core/src/metadata/ng_module';
import {defineInjector, InjectorType} from '../packages/angular/core/src/di/defs';


export function makeDecorator(
  name: string, props?: (...args: any[]) => any, parentClass?: any,
  chainFn?: (fn: Function) => void, typeFn?: (type: Type<any>, ...args: any[]) => void): {new (...args: any[]): any; (...args: any[]): any; (...args: any[]): (cls: any) => any;} {
  const metaCtor = makeMetadataCtor(props);

  function DecoratorFactory(...args: any[]): (cls: any) => any {
    console.log(this);

    if (this instanceof DecoratorFactory) {
      metaCtor.call(this, ...args);
      return this;
    }

    const annotationInstance = new (<any>DecoratorFactory)(...args);
    const TypeDecorator: TypeDecorator = <TypeDecorator>function TypeDecorator(cls: Type<any>) {
      typeFn && typeFn(cls, ...args);
      // Use of Object.defineProperty is important since it creates non-enumerable property which
      // prevents the property is copied during subclassing.
      const annotations = cls.hasOwnProperty(ANNOTATIONS) ?
        (cls as any)[ANNOTATIONS] :
        Object.defineProperty(cls, ANNOTATIONS, {value: []})[ANNOTATIONS];
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
  (<any>DecoratorFactory).annotationCls = DecoratorFactory;


  console.log(DecoratorFactory);

  return DecoratorFactory as any;
}



function makeMetadataCtor(props?: (...args: any[]) => any): any {
  return function ctor(...args: any[]) {
    if (props) {
      console.log(...args, );

      const values = props(...args);
      for (const propName in values) {
        this[propName] = values[propName];
      }
    }
  };
}



///////////////////////////////////////////////////////////////////

export interface NgModuleDecorator {
  /**
   * Defines an NgModule.
   */
  (obj?: NgModule): TypeDecorator;
  new (obj?: NgModule): NgModule;
}

export interface NgModule {
  providers?: Provider[];
  declarations?: Array<Type<any>|any[]>;
  imports?: Array<Type<any>|ModuleWithProviders|any[]>;
  exports?: Array<Type<any>|any[]>;
  entryComponents?: Array<Type<any>|any[]>;
  bootstrap?: Array<Type<any>|any[]>;
  schemas?: Array<SchemaMetadata|any[]>;
  id?: string;
}


export const NgModule: NgModuleDecorator = makeDecorator('NgModule', (ngModule: NgModule) => ngModule, undefined, undefined,
  (moduleType: InjectorType<any>, metadata: NgModule) => {
    let imports = (metadata && metadata.imports) || [];
    if (metadata && metadata.exports) {
      imports = [...imports, metadata.exports];
    }

    moduleType.ngInjectorDef = defineInjector({
      factory: convertInjectableProviderToFactory(moduleType, {useClass: moduleType}),
      providers: metadata && metadata.providers,
      imports: imports,
    });
  });


@NgModule({
  providers: [
    {provide: 'a', useValue: 'a'}
  ]
})
export class AppModule {
  
}

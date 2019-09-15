
import {
  AfterViewInit, Component,
  Directive, ElementRef,
  EventEmitter,
  Inject,
  InjectionToken,
  Input, NgModule, OnChanges,
  Optional,
  Output, Renderer2, SimpleChanges
} from "@angular/core";
import {BrowserModule} from "@angular/platform-browser";


interface Validator {

}

const NG_VALIDATORS = new InjectionToken<Validator[]>('NgValidators');

interface ValidationErrors {
  [key: string]: any;
}

interface ValidatorFn {
  (control: FormControl): ValidationErrors | null;
}


class FormGroup {
  constructor(public controls: {[key: string]: FormControl}, validator?: ValidatorFn) {

  }
}

class FormControl {
  value: any;

  setValue(newValue: any) {

  }
}


class ControlContainer {

}

/**
 * Create a top-level 'FormGroup' instance
 */
@Directive({
  selector: 'form:not([ngNoForm])',
  host: {'(submit)': 'onSubmit($event)', '(reset)': 'onReset($event)'},
  exportAs: 'ngForm'
})
export class Form extends ControlContainer implements AfterViewInit {
  @Output() submit = new EventEmitter();

  submitted = false;
  form: FormGroup;


  constructor(@Optional() @Inject(NG_VALIDATORS) validators: Validator[]) {
    super();

    this.form = new FormGroup({});
  }

  onSubmit($event: Event) {
    this.submitted = true;
    this.submit.emit($event);
  }

  onReset($event: Event) {
    this.submitted = false;

  }

  ngAfterViewInit() {

  }
}



interface ControlValueAccessor {
  writeValue(value: any): void;
  registerOnChange(fn: any): void;
}


abstract class NgControl {
  valueAccessor: ControlValueAccessor | null;

  abstract viewToModelUpdate(newValue: any): void;
}



function setUpControl(control: FormControl, directive: NgControl) {
  directive.valueAccessor!.writeValue(control.value);

  directive.valueAccessor!.registerOnChange((newValue) => {
    control.setValue(newValue);
    directive.viewToModelUpdate(newValue);
  });
}

/**
 * Create a 'FormControl' instance and bind it to a DOM element.
 * If DOM element value changes, 'FormControl' value field changes auto.
 * If 'FormControl' value field changes, DOM element value changes auto.
 */
@Directive({
  selector: '[ngModel]',
  exportAs: 'ngModel'
})
export class Model extends NgControl implements OnChanges {
  @Input('ngModel') model: any;
  @Output() ngModelChange = new EventEmitter();

  registered = false;
  control = new FormControl();

  constructor(@Optional() parent: ControlContainer, @Optional() @Inject(NG_VALIDATORS) validators: Validator[]) {
    super();
  }


  ngOnChanges(changes: SimpleChanges) {
    console.log(changes);

    if (! this.registered) {
      // bind FormControl instance to DOM element
      setUpControl(this.control, this);
    }


    this.registered = true;
  }

  viewToModelUpdate(value) {
    this.ngModelChange.emit(value);
  }
}



@Directive({
  selector: 'input[type=text]',
  host: {
    '(input)':'handleInput($event.target.value)'
  }
})
export class InputControlValueAccessor implements ControlValueAccessor {
  onChange: (value) => void;

  constructor(private _render: Renderer2, private _element: ElementRef) {}


  handleInput(value) {
    this.onChange(value);
  }

  registerOnChange(fn: (value) => void) {
    this.onChange = fn;
  }

  writeValue(value: any) {
    this._render.setProperty(this._element, 'value', value);
  }
}





@Component({
  selector: 'demo-forms',
  template: `
    <h2>Test Bidirectional Data Binding</h2>
    <p>{{name}}</p>
    <button (click)="name='new name'">Change Name</button>
    <input [ngModel]="name" (ngModelChange)="name=$event.target.value" (input)="name=$event.target.value" [value]="name"/>
    
    <h2>NgModel</h2>
    <p>{{name2}}</p>
    <button (click)="name2='new name2'">Change Name2</button>
    <input type='text' [ngModel]="name2" (ngModelChange)="name2=$event.target.value"/>
    
    <p>{{age}}</p>
    <button (click)="age=11">Change Age</button>
    <select (change)="age=$event.target.value">
      <option [value]="10">1</option>
      <option [value]="20">2</option>
      <option [value]="30">3</option>
    </select>
    
    <p>{{checked}}</p>
    <button (click)="checked=!checked">Change Checked Status</button>
    <input type="checkbox" (change)="checked=$event.target.value" [checked]="checked"/>
  `
})
export class AppComponent {
  name = 'old name';
  name2 = 'old name2';
  age = 1;
  checked = false;
}


@NgModule({
  imports: [
    BrowserModule,
  ],
  declarations: [
    AppComponent,
    Model,
  ],
  bootstrap: [AppComponent]
})
export class DemoFormsModule {

}









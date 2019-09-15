

/**
 * Three Ways to Dynamically Alter your Form Validation in Angular
 * @see https://netbasal.com/three-ways-to-dynamically-alter-your-form-validation-in-angular-e5fd15f1e946
 *
 *
 * Design Docs: https://docs.google.com/document/d/1dlJjRXYeuHRygryK0XoFrZNqW86jH4wobftCFyYa1PA/edit
 *
 * Destination: FormGroupDirective, FormControlName
 */


import {Component, ErrorHandler, NgModule, OnInit} from "@angular/core";
import {BrowserModule} from "@angular/platform-browser";
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validator, Validators} from "@angular/forms";

@Component({
  selector: 'alter-form-validation',
  template: `
    <form [formGroup]="form">

      <input type="checkbox" formControlName="optionA"> Option A
      <input type="checkbox" formControlName="optionB"> Option B

      <input formControlName="optionBExtra" placeholder="Reason" *ngIf="optionBExtra">

    </form>

    <p>Form Value: {{ form.value | json }}</p>
    <p>Form Valid: {{ form.valid }}</p>
  `
})
export class AppComponent implements OnInit {
  form: FormGroup;
  
  ngOnInit() {
    this.form = new FormGroup({
      optionA: new FormControl(),
      optionB: new FormControl(),
    });
    
    
    this.optionB.valueChanges.subscribe((checked) => {
      console.log(checked);
      
      if (checked) {
        const validators = [Validators.required, Validators.minLength(5)];
        this.form.addControl('optionBExtra', new FormControl('', validators));
      } else {
        this.form.removeControl('optionBExtra');
      }
      // this.form.updateValueAndValidity();
    });
  }
  
  get optionB() {
    return this.form.get('optionB') as FormControl;
  }
  
  get optionBExtra() {
    return this.form.get('optionBExtra') as FormControl;
  }
}


export class SentryHandler extends ErrorHandler {
  handleError(error: any) {
  
  }
}


@NgModule({
  imports: [
    BrowserModule,
    ReactiveFormsModule,
  ],
  declarations: [
    AppComponent,
  ],
  providers: [
    // {provide: ErrorHandler, useClass: SentryHandler}
  ],
  bootstrap: [
    AppComponent,
  ]
})
export class FormValidationModule {

}
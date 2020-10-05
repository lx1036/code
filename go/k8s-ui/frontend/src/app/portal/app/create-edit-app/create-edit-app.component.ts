import {Component, EventEmitter, OnInit, Output, ViewChild} from "@angular/core";
import {NgForm} from "@angular/forms";
import {ActionType} from "../app.component";
import {App} from "../../../shared/models/app";
import {AppService} from "../../../shared/common/app.service";


@Component({
  selector: 'app-create-edit-app',
  templateUrl: './create-edit-app.component.html',
})
export class CreateEditAppComponent implements OnInit {
  createAppOpened: boolean;
  @Output() create = new EventEmitter<boolean>();
  appTitle: string;
  isNameValid: boolean;
  app: App = new App();
  actionType: ActionType;
  @ViewChild('appForm', { static: true }) currentForm: NgForm;
  checkOnGoing = false;
  isSubmitOnGoing = false;

  constructor(private appService: AppService, ) {}

  ngOnInit() {
  }

  public get isValid(): boolean {
    return this.currentForm &&
      this.currentForm.valid &&
      !this.isSubmitOnGoing &&
      this.isNameValid &&
      !this.checkOnGoing;
  }

  newOrEditApp(id?: number) {

  }

  onCancel() {

  }

  handleValidation() {

  }

  onSubmit() {
    switch (this.actionType) {
      case ActionType.ADD_NEW:
        this.appService.create(this.app).subscribe(response => {}, err => {});
        break;
      case ActionType.EDIT:
        this.appService.update(this.app).subscribe(response => {}, err => {});
        break;
    }
  }
}

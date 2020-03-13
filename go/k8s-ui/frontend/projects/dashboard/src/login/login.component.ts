import {Component, OnInit} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {ActivatedRoute} from "@angular/router";
import {KdError, StateError} from "../typings/api";
import {map} from "rxjs/operators";


enum LoginModes {
  Kubeconfig = 'kubeconfig',
  Basic = 'basic',
  Token = 'token',
}

@Component({
  selector: 'kube-login',
  template: `
    <div class="kd-login-container kd-bg-background" fxFlex>
      <kube-card titleClasses="kd-card-top-radius kd-bg-primary kd-accent" class="kd-login-card" [expandable]="false">
        <div title i18n>Kubernetes Dashboard</div>
        <div content>
          <form fxLayout="column" (ngSubmit)="login()">
            <mat-radio-group name="login" [(ngModel)]="selectedAuthenticationMode">
              <div *ngFor="let mode of getEnabledAuthenticationModes()">
                <mat-radio-button [value]="mode" color="primary">
                  <ng-container [ngSwitch]="mode">
                    <ng-container *ngSwitchCase="loginModes.Kubeconfig" i18n>Kubeconfig</ng-container>
                    <ng-container *ngSwitchCase="loginModes.Basic" i18n>Basic</ng-container>
                    <ng-container *ngSwitchCase="loginModes.Token" i18n>Token</ng-container>
                  </ng-container>
                </mat-radio-button>
                <div class="kd-login-mode-description" [ngSwitch]="mode">
                  <ng-container *ngSwitchCase="loginModes.Kubeconfig" i18n>
                    Please select the kubeconfig file that you have created to configure access to the cluster. To find out more about how to configure and use kubeconfig file, please refer to the <a href='https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/'>Configure Access to Multiple Clusters</a> section.
                  </ng-container>
                  <ng-container *ngSwitchCase="loginModes.Basic" i18n>
                    Make sure that support for basic authentication is enabled in the cluster. To find out more about how to configure basic authentication, please refer to the <a href="https://kubernetes.io/docs/admin/authentication/">Authenticating</a> and <a href="https://kubernetes.io/docs/admin/authorization/abac/">ABAC Mode</a> sections.
                  </ng-container>
                  <ng-container *ngSwitchCase="loginModes.Token" i18n>
                    Every Service Account has a Secret with valid Bearer Token that can be used to log in to Dashboard. To find out more about how to configure and use Bearer Tokens, please refer to the <a href='https://kubernetes.io/docs/admin/authentication/'>Authentication</a> section.
                  </ng-container>
                </div>
              </div>
            </mat-radio-group>

            <ng-container [ngSwitch]="selectedAuthenticationMode">
              <mat-form-field *ngSwitchCase="loginModes.Token" class="kd-login-input">
                <input matInput id="token" name="token" i18n-placeholder placeholder="Enter token" type="password" required (change)="onChange($event)">
              </mat-form-field>
              <div *ngSwitchCase="loginModes.Basic" fxLayout="column">
                <mat-form-field fxFlex class="kd-login-input">
                  <input id="username" name="username" matInput i18n-placeholder placeholder="Username" required (change)="onChange($event)">
                </mat-form-field>
                <mat-form-field fxFlex class="kd-login-input">
                  <input id="password" name="password" matInput i18n-placeholder placeholder="Password" type="password" required (change)="onChange($event)">
                </mat-form-field>
              </div>

              <div *ngSwitchCase="loginModes.Kubeconfig" class="kd-login-input">
                <kd-upload-file label="Choose kubeconfig file" i18n-label (onLoad)="onChange($event)"></kd-upload-file>
              </div>
              <ng-template ngFor let-error [ngForOf]="errors" ngProjectAs="mat-error" class="kd-login-input">
                <mat-error class="kd-login-input kd-error kd-error-text">
                  {{error.status}} ({{error.code}}): {{error.message}}
                </mat-error>
              </ng-template>
            </ng-container>

            <div fxFlex="none" fxLayout="row">
              <button mat-raised-button color="primary" type="submit" class="kd-login-button" i18n>
                Sign in
              </button>
              <button mat-button color="primary" type="button" class="kd-login-button" *ngIf="isSkipButtonEnabled()" (click)="skip()" i18n>
                Skip
              </button>
            </div>
          </form>
        </div>
      </kube-card>
    </div>
  `
})
export class LoginComponent implements OnInit {
  errors: KdError[] = [];

  constructor(private readonly http: HttpClient,private readonly route: ActivatedRoute,) {
  }

  ngOnInit() {
    this.http
    .get<EnabledAuthenticationModes>('api/v1/login/modes')
    .subscribe((enabledModes: EnabledAuthenticationModes) => {
      this.enabledAuthenticationModes_ = enabledModes.modes;
    });

    this.http
    .get<LoginSkippableResponse>('api/v1/login/skippable')
    .subscribe((loginSkippableResponse: LoginSkippableResponse) => {
      this.isLoginSkippable_ = loginSkippableResponse.skippable;
    });

    this.route.paramMap.pipe(map(() => window.history.state)).subscribe((state: StateError) => {
      if (state.error) {
        this.errors = [state.error];
      }
    });
  }
}


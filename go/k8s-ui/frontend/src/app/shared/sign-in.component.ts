import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgForm} from '@angular/forms';
import {AuthoriseService} from './client/v1/auth.service';
import {AuthService} from './auth.service';
import * as particlesJS from 'particlesjs/dist/particles';
import {AuthType} from './shared.const';

@Component({
  selector: 'app-sign-in',
  template: `
    <div class="container">
      <form class="login" #ngForm="ngForm">
        <div>
          <div style="padding-bottom:5px;text-align:center;">
            <img src="assets/images/wayne-logo.blue.svg" width="200px" alt=""><br>
            <strong>{{getTitle()}} {{version}}</strong>
          </div>
        </div>

        <div *ngIf="authService.config?.enableDBLogin || authService.config?.ldapLogin"
             style="padding-bottom:5px;text-align:center;">
          <div>
            <app-input style="margin-top: 20px;height: 35px;font-size: 14px;" required [(ngModel)]="username" type="text"
                         name="login_username" id="login_username" placeholder="请输入用户名"></app-input>
            <app-input style="margin-top: 20px;height: 35px;font-size: 14px;" required [(ngModel)]="password"
                         type="password" name="login_password" id="login_password" placeholder="请输入密码"></app-input>
          </div>
          <div *ngIf="errMsg" class="error">{{errMsg}}</div>
          <div>
            <button style="margin-top: 20px; width:240px;height: 40px;font-size: 16px;" type="submit" class="wayne-button"
                    [class.normal]="isValid" [class.invalid]="!isValid" (click)="onSubmit()">立即登录
            </button>
          </div>
        </div>

        <div *ngIf="authService.config?.oauth2Login" style="padding-bottom:5px;text-align:center;">
          <hr style="width: 240px;"/>
          <button style="margin-top: 10px; width:240px;height: 40px;font-size: 16px;" type="submit" (click)="oauth2Login()"
                  class="wayne-button normal">{{getOAuth2Title()}}</button>
        </div>
      </form>
    </div>
    <canvas class="background"></canvas>
  `,
  styles: [
    `
      .container {
        width: 332px;
        max-height: 453px;
        overflow: auto;
        margin-top: 10%;
        margin-left: auto;
        margin-right: 15%;
        padding: 10px 50px 20px 50px;
        background-color: #fff;
        box-shadow: 0 0 24px 0 rgba(0, 0, 0, 0.06), 0 1px 0 0 rgba(0, 0, 0, 0.02);
      }

      .background {
        position: absolute;
        top: 0;
        bottom: 0;
        left: 0;
        right: 0;
        z-index: -1;
        background-color: #F4F7FB;
      }

      .error {
        color: #FF3434;
        font-size: 12px;
      }
    `
  ]
})
export class SignInComponent implements OnInit {
  version = require('../../../package.json').version;
  ngForm: NgForm;
  errMsg: string;
  username: string;
  password: string;
  currentForm: NgForm;
  isSubmitOnGoing: boolean;

  constructor(private authoriseService: AuthoriseService,
              private route: ActivatedRoute,
              public authService: AuthService) {
  }

  ngOnInit() {
    particlesJS.init({ // add particles in background, beautiful!
      selector: '.background',
      color: ['#DA0463', '#404B69', '#DBEDF3'],
      connectParticles: true
    });
  }

  public get isValid(): boolean {
    return this.currentForm &&
      this.currentForm.valid &&
      !this.isSubmitOnGoing;
  }

  getTitle() {
    const imagePrefix = this.authService.config['system.title'];
    return imagePrefix ? imagePrefix : 'Kubernetes-UI';
  }

  getOAuth2Title() {
    const oauth2Title = this.authService.config['system.oauth2-title'];
    return oauth2Title ? oauth2Title : 'OAuth 2.0 Login';
  }

  onSubmit() {
    let type = AuthType.DB;
    if (this.authService.config && this.authService.config.ldapLogin) {
      type = AuthType.Ldap;
    }
    
    this.authoriseService.login(this.username, this.password, type).subscribe(
      (response) => {
        const refer = this.route.snapshot.queryParams['ref'] ? this.route.snapshot.queryParams['ref'] : '/';
        
      },
      (error) => {}
    )
  }

  oauth2Login() {

  }
}

import {Injectable, Injector} from '@angular/core';
import {MessageService} from '../global-message/message.service';
import {AlertType, httpStatusCode} from '../shared.const';
import {HttpClient} from '@angular/common/http';
import {Router} from '@angular/router';

@Injectable({
  providedIn: 'root'
})
export class MessageHandlerService {

  constructor(private msgService: MessageService, private injector: Injector) { }

  public showError(message: string): void {
    if (message && message.trim() !== '') {
      this.msgService.announceMessage(500, message, AlertType.DANGER);
    }
  }




  error(error: string) {
    this.showError(error);

  }

  handleError(error: any) {
    if (!error) {
      return;
    }
    const code = error.statusCode || error.status;

    if (code === httpStatusCode.Unauthorized) {
      const currentUrl = document.location.origin;
      if (document.location.pathname !== '/sign-in') {
        this.injector.get(Router).navigateByUrl(`sign-in?ref=${document.location.pathname}`);
      }
    } else {
      this.msgService.announceMessage(code, error.error ? error.error.msg : error.error, AlertType.DANGER);
    }
  }
}

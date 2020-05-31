

/**
 * This file includes polyfills needed by Angular and is loaded before the app.
 * You can add your own extra polyfills to this file.
 */

// IE10 and IE11 requires the following for the Reflect API.
import 'core-js/es/reflect';

// Required to support Web Animations `@angular/platform-browser/animations`:
import 'web-animations-js';

// Zone JS is required by default for Angular itself.
import 'zone.js/dist/zone';

// RxJS is required to support additional Observable methods such as map or switchMap.
import 'rxjs/Rx';

// Load `$localize` onto the global scope - used if i18n tags appear in Angular templates.
import '@angular/localize/init';

/* tslint:disable */
// Global variable is required by some 3rd party libraries such as 'ace-ui'.
// It was removed in Angular 6.X, more info can be found here:
// https://github.com/angular/angular-cli/issues/9827#issuecomment-369578814
(window as any).global = window;

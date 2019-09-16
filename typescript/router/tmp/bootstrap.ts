import {MainModuleNgFactory} from './test.ngfactory';
import {platformBrowser} from '@angular/platform-browser';

const platform = platformBrowser().bootstrapModuleFactory(MainModuleNgFactory);

console.log(platform);

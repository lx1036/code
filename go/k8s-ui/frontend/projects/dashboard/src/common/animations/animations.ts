

import {animate, AUTO_STYLE, state, style, transition, trigger} from '@angular/animations';

const DEFAULT_TRANSITION_TIME = '500ms ease-in-out';

export class Animations {
  static easeOut = trigger('easeOut', [
    transition('* => void', [
      style({opacity: 1}),
      animate(DEFAULT_TRANSITION_TIME, style({opacity: 0})),
    ]),
  ]);

  static easeInOut = trigger('easeInOut', [
    transition('void => *', [
      style({opacity: 0}),
      animate(DEFAULT_TRANSITION_TIME, style({opacity: 1})),
    ]),
    transition('* => void', [animate(DEFAULT_TRANSITION_TIME, style({opacity: 0}))]),
  ]);

  static expandInOut = trigger('expandInOut', [
    state('true', style({height: '0', display: 'none'})),
    state(
      'false',
      style({
        height: AUTO_STYLE,
        display: AUTO_STYLE,
      }),
    ),
    transition('false => true', [
      style({overflow: 'hidden'}),
      animate('500ms ease-in', style({height: '0'})),
    ]),
    transition('true => false', [
      style({overflow: 'hidden'}),
      animate('500ms ease-out', style({height: AUTO_STYLE})),
    ]),
  ]);
}

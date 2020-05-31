

import {Pipe, SecurityContext} from '@angular/core';
import {DomSanitizer, SafeHtml} from '@angular/platform-browser';

const ansiColorClass = require('ansi-to-html');
const ansiColor = new ansiColorClass();

enum TextMode {
  Default = 'Default',
  Colored = 'Colored',
}

/**
 * Formats the given value as raw HTML to display to the user.
 */
@Pipe({name: 'kdSafeHtml'})
export class SafeHtmlFormatter {
  constructor(private readonly sanitizer: DomSanitizer) {}

  transform(value: string, mode: TextMode = TextMode.Default): SafeHtml {
    let result: SafeHtml = null;
    let content = this.sanitizer.sanitize(
      SecurityContext.HTML,
      value.replace('<', '&lt;').replace('>', '&gt;'),
    );

    // Handle conversion of ANSI color codes.
    switch (mode) {
      case TextMode.Colored:
        content = ansiColor.toHtml(content.replace(/&#27;/g, '\u001b'));
        result = this.sanitizer.bypassSecurityTrustHtml(content);
        break;

      default:
        // TextMode.Default
        result = content;
        break;
    }

    return result;
  }
}

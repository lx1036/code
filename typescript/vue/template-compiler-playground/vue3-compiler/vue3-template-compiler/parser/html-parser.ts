import {makeMap} from "../utils";

const comment = /^<!\--/;
const doctype = /^<!DOCTYPE [^>]+>/i;
const unicodeRegExp = /a-zA-Z\u00B7\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u037D\u037F-\u1FFF\u200C-\u200D\u203F-\u2040\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD/;
const ncname = `[a-zA-Z_][\\-\\.0-9_a-zA-Z${unicodeRegExp.source}]*`;
const qnameCapture = `((?:${ncname}\\:)?${ncname})`;
const endTag = new RegExp(`^<\\/${qnameCapture}[^>]*>`);
const startTagOpen = new RegExp(`^<${qnameCapture}`);
const startTagClose = /^\s*(\/?)>/;
const dynamicArgAttribute = /^\s*((?:v-[\w-]+:|@|:|#)\[[^=]+\][^\s"'<>\/=]*)(?:\s*(=)\s*(?:"([^"]*)"+|'([^']*)'+|([^\s"'=<>`]+)))?/;
const attribute = /^\s*([^\s"'<>\/=]+)(?:\s*(=)\s*(?:"([^"]*)"+|'([^']*)'+|([^\s"'=<>`]+)))?/;


const isPlainTextElement = makeMap('script,style,textarea', true);


type StartTagMatch = {
  tagName: string, // ["<div", "div", index: 0, input: "<div class="box"></div>", groups: undefined]
  attrs: [],
  start: number,
  unarySlash: string, // 自闭合标签则为 '/'
  end: number,
}

/**
 * Parse html:
 * <div>
 *   <p>{{name}}</p>
 * </div>
 * to:
 *
 * @param html
 * @param options
 */



export const parseHTML = (html: string, options: any) => {
  let last: string, lastTag: string;
  let index = 0;
  const expectHTML = options.expectHTML;
  
  /**
   *
   */
  const advance = (step: number) => {
    index += step;
    html = html.substring(step);
  };
  
  /**
   *
   */
  const parseEndTag = (tagName: string, start: number, end: number) => {
    let pos, lowerCasedTagName;
    
    if (tagName) {
      lowerCasedTagName = tagName.toLowerCase();
      
    }
    
  };
  
  /**
   *
   */
  const parseStartTag = () => {
    const start = html.match(startTagOpen); // <div class="box"></div>
  
    if (start) {
      let end, attr: {start: number, end: number};
      const match: StartTagMatch = {
        tagName: start[1], // ["<div", "div", index: 0, input: "<div class="box"></div>", groups: undefined]
        attrs: [],
        start: index,
        unarySlash: '', // 自闭合标签则为 '/'
        end: 0,
      };
      
      advance(start[0].length); // 截取 <div 后为 ' class="box></div>"'
      
      /**
       *
       * 依次截取 '<div class="box" id="name"></div>' attr 属性，把
       */
      // ' class="box" id="name"></div>'.match(attribute) -> attr=[" class="box"", "class", "=", "box", undefined, undefined, index: 0, input: " class="box" id="name"></div>", groups: undefined]
      while (!(end = html.match(startTagClose)) && (attr = html.match(dynamicArgAttribute) || html.match(attribute))) {
        attr.start = index;
        advance(attr[0].length);
        attr.end = index;
        match.attrs.push(attr);
      }
    
      // startTagClose 开始标签的闭合部分
      if (end) { // [">", "", index: 0, input: "></div>", groups: undefined] or 自闭合标签 ["/>", "/", index: 0, input: "/>", groups: undefined]
        match.unarySlash = end[1];
        advance(end[0].length);
        match.end = index;
      
        return match;
      }
    }
  
    return;
  };
  
  const handleEndTag = () => {};
  
  /**
   * 将 tagName,attrs,unary 等数据取出来，并调用钩子函数把数据存入参数中
   */
  const handleStartTag = (match) => {
    const tagName = match.tagName;
    const unarySlash = match.unarySlash;
  
    const l = match.attrs.length;
    const attrs = new Array(l);
  
    for (let i = 0; i < l; i++) {
    
    }
  
    if (options.start) {
      options.start(tagName, attrs, !!unarySlash, match.start, match.end);
    }
  };
  
  
  
  while (html) {
    last = html;
    
    // Make sure we're not in a plaintext content element like script/style
    if (!lastTag || !isPlainTextElement(lastTag)) {
      let textEnd = html.indexOf('<');
      
      if (textEnd === 0) { // 起始字符是 '<'
        // Comment: <!--<h1>This is an about page</h1>-->
        if (comment.test(html)) {
          const commentEnd = html.indexOf('-->');
          
          if (commentEnd >= 0) {
          
            advance(commentEnd + 3);
            continue;
          }
        }
        
        // <!DOCTYPE html>
        const doctypeMatch = html.match(doctype);
        if (doctypeMatch) {
          advance(doctypeMatch[0].length);
          continue;
        }
        
        // End Tag
        const endTagMatch = html.match(endTag);
        if (endTagMatch) {
          const start = index;
          advance(endTagMatch[0].length);
          parseEndTag(endTagMatch[1], start, index);
          continue;
        }
        
        // Start Tag
        const startTagMatch = parseStartTag();
        if (startTagMatch) {
          handleStartTag(startTagMatch);
          
          continue;
        }
      }
      
      let text;
      
      if (textEnd >= 0) { // 文本字符串里有 '<'
      
      }
      
      if (textEnd < 0) { // html 是文本
        text = html
      }
      
      if (text) {
        advance(text.length);
      }
    } else { // 纯文本内容元素: is the <template> or <script> or <style> tag
      const stackedTag = lastTag.toLowerCase();
      
      // parseEndTag(stackedTag);
    }
    
    if (last === html) {
      break;
    }
  }
  
  
  
  
  
};


/**
 * Demo:
 */

// const htmlParser = new HtmlParser();
// htmlParser.parse('<template><div><p>{{name}}</p></div></template>', {});



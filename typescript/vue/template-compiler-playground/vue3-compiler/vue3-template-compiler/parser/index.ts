import {ASTElement, CompilerOptions} from '../types';
import {parseHTML} from "./html-parser";


function makeAttributesMap(attributes: Array<any>) {
  const map: any = {};
  
  for (let i = 0, l = attributes.length; i < l; i++) {
    map[attributes[i].name] = attributes[i].value;
  }
  
  return map;
}

function createASTElement(tag: string, attributes: Array<any>, parent: ASTElement) {
  return {
    type: 1,
    tag,
    attributesList: attributes,
    attributesMap: makeAttributesMap(attributes),
    rawAttributesMap: {},
    parent,
    children: []
  }
}



/**
 * Convert HTML string to AST.
 */
export const parse = (template: string, options: any): ASTElement => {
  let root: any;
  let currentParent: any;
  
  parseHTML(template, {
    expectHTML: options.expectHTML,
    start(tag: any, attributes: Array<any>) {
      let element = createASTElement(tag, attributes, currentParent);
      
      if (!root) {
        root = element;
      }
    }
  });
  
  return root;
};

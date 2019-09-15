

export class VNode {
  constructor(tagName, attributes, children, componentOptions, componentInstance) {
    this.tagName = tagName;
    this.attributes = attributes;
    this.children = children;
    this.componentOptions = componentOptions;
    this.componentInstance = componentInstance;
  }
}

export function createTextNode(value) {
  return new VNode(undefined, undefined, value);
}

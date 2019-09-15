export type ASTAttr = {
  name: string;
  value: any;
  dynamic?: boolean;
  start?: number;
  end?: number
};

interface ModuleOptions {
  preTransformNode: (el: ASTElement) => ASTElement | undefined;
  transformNode: (el: ASTElement) => ASTElement | undefined;
  postTransformNode: (el: ASTElement) => void;
  genData: (el: ASTElement) => string;
  transformCode?: (el: ASTElement, code: string) => string;
  staticKeys?: string[];
}

export interface ASTIfCondition {
  exp: string | undefined;
  block: ASTElement;
}

export interface ASTElementHandler {
  value: string;
  params?: any[];
  modifiers: ASTModifiers | undefined;
}

export interface ASTElementHandlers {
  [key: string]: ASTElementHandler | ASTElementHandler[];
}

export interface ASTModifiers {
  [key: string]: boolean;
}

export interface ASTDirective {
  name: string;
  rawName: string;
  value: string;
  arg: string | undefined;
  modifiers: ASTModifiers | undefined;
}
type DirectiveFunction = (node: ASTElement, directiveMeta: ASTDirective) => void;

export interface ASTExpression {
  type: 2;
  expression: string;
  text: string;
  tokens: (string | Record<string, any>)[];
  static?: boolean;
  // 2.4 ssr optimization
  // ssrOptimizability?: SSROptimizability;
}

export interface ASTText {
  type: 3;
  text: string;
  static?: boolean;
  isComment?: boolean;
  // 2.4 ssr optimization
  // ssrOptimizability?: SSROptimizability;
}


export interface ASTElement {
  type: 1;
  tag: string;
  attrsList: { name: string; value: any }[];
  attrsMap: Record<string, any>;
  parent: ASTElement | undefined;
  children: ASTNode[];
  
  processed?: true;
  
  static?: boolean;
  staticRoot?: boolean;
  staticInFor?: boolean;
  staticProcessed?: boolean;
  hasBindings?: boolean;
  
  text?: string;
  attrs?: { name: string; value: any }[];
  props?: { name: string; value: string }[];
  plain?: boolean;
  pre?: true;
  ns?: string;
  
  component?: string;
  inlineTemplate?: true;
  transitionMode?: string | null;
  slotName?: string;
  slotTarget?: string;
  slotScope?: string;
  scopedSlots?: Record<string, ASTElement>;
  
  ref?: string;
  refInFor?: boolean;
  
  if?: string;
  ifProcessed?: boolean;
  elseif?: string;
  else?: true;
  ifConditions?: ASTIfCondition[];
  
  for?: string;
  forProcessed?: boolean;
  key?: string;
  alias?: string;
  iterator1?: string;
  iterator2?: string;
  
  staticClass?: string;
  classBinding?: string;
  staticStyle?: string;
  styleBinding?: string;
  events?: ASTElementHandlers;
  nativeEvents?: ASTElementHandlers;
  
  transition?: string | true;
  transitionOnAppear?: boolean;
  
  model?: {
    value: string;
    callback: string;
    expression: string;
  };
  
  directives?: ASTDirective[];
  
  forbidden?: true;
  once?: true;
  onceProcessed?: boolean;
  wrapData?: (code: string) => string;
  wrapListeners?: (code: string) => string;
  
  // 2.4 ssr optimization
  // ssrOptimizability?: SSROptimizability;
  
  // weex specific
  appendAsTree?: boolean;
}

export interface CompilerOptions {
  modules?: ModuleOptions[];
  directives?: Record<string, DirectiveFunction>;
  preserveWhitespace?: boolean;
  whitespace?: 'preserve' | 'condense';
  outputSourceRange?: any
}

export interface CompiledResult<ErrorType> {
  ast: ASTElement | undefined;
  render: string;
  staticRenderFns: string[];
  errors: ErrorType[];
  tips: ErrorType[];
}




export type ASTNode = ASTElement | ASTText | ASTExpression;

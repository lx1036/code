import {baseOptions} from "../platform-web-compiler/options";
import {parse} from "../vue3-template-compiler/parser";

describe('parser', () => {
  it('simple element', () => {
    const ast = parse('<h1>hello world</h1>', baseOptions);
    expect(ast.tag).toBe('h1');
    expect(ast.plain).toBe(true);
    expect(ast.children[0].text).toBe('hello world');
  });
  
  it('interpolation in element', () => {
    const ast = parse('<h1>{{msg}}</h1>', baseOptions);
    expect(ast.tag).toBe('h1');
    expect(ast.plain).toBe(true);
    expect(ast.children[0].expression).toBe('_s(msg)');
  });
  
  it('child elements', () => {
    const ast = parse('<ul><li>hello world</li></ul>', baseOptions);
    expect(ast.tag).toBe('ul');
    expect(ast.plain).toBe(true);
    expect(ast.children[0].tag).toBe('li');
    expect(ast.children[0].plain).toBe(true);
    expect(ast.children[0].children[0].text).toBe('hello world');
    expect(ast.children[0].parent).toBe(ast);
  });
  
  it('unary element', () => {
    const ast = parse('<hr>', baseOptions);
    expect(ast.tag).toBe('hr');
    expect(ast.plain).toBe(true);
    expect(ast.children.length).toBe(0);
  });
  
  it('camelCase element', () => {
    const ast = parse('<MyComponent><p>hello world</p></MyComponent>', baseOptions)
    expect(ast.tag).toBe('MyComponent')
    expect(ast.plain).toBe(true)
    expect(ast.children[0].tag).toBe('p')
    expect(ast.children[0].plain).toBe(true)
    expect(ast.children[0].children[0].text).toBe('hello world')
    expect(ast.children[0].parent).toBe(ast)
  })
});

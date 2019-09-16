







//
// export function createCompilerCreator(base: Function): Function {
//
// }





/*
 * Template compilation options / results
 */
import {ASTElement, CompiledResult, CompilerOptions} from "./types";
import {parse} from "./parser";
import {optimize} from "./optimizer";
import {generate} from "./codegenerator";

const createCompileToFunction = (compile) => {};

const createCompilerCreator = (baseCompile: Function) => {
  return (options: any) => {
    const compile = (template: any, options: any) => {
      const compiled = baseCompile(template, options);
      
      return compiled;
    };
    
    return {compile, compileToFunctions: createCompileToFunction(compile)};
  };
};

export const createCompiler = createCompilerCreator((template: string, options: CompilerOptions): CompiledResult<ErrorType> => {
  const ast = parse(template.trim(), options);
  optimize(ast, options);
  const code = generate(ast, options);
  
  return {code, render: code.render, staticRenderFns: code.staticRenderFns};
});










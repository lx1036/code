import {createCompiler} from "../vue3-template-compiler";
import {baseOptions} from "./options";


export const {compile, compileToFunctions} = createCompiler(baseOptions);

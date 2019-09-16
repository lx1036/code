/**
 * ../../../node_modules/.bin/jest -c ./jest.config.js
 * ../../../node_modules/.bin/webpack-dev-server ./index.ts --config ./webpack.config.js
 *
 * @see https://github.com/vuejs/vue/blob/master/src/platforms/web/compiler/index.js
 */

// import {compile, CompiledResult} from 'vue-template-compiler';
import {compile, CompiledResult} from './platform-web-compiler';


let template = `<div class="app"><p class="title" v-if="visible">Title</p></div>`;
let compiledResult: CompiledResult<string> = compile(template);

console.log(compiledResult.ast, compiledResult.render, compiledResult.staticRenderFns);

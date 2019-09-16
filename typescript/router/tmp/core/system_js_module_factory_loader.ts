
// import * as _  from 'systemjs';
// // import SystemJSLoader = require("systemjs");
// import {TIME} from "./modules/hello.module";
//
// // _.config({transpiler: 'typescript', paths: {
// //     'typescript': './node_modules/typescript',
// //   },
// //   packages: {
// //     'typescript': {
// //       main: 'lib/typescript'
// //     },
// //     'server': {
// //       defaultExtension: 'ts'
// //     }
// //   },
// // });
//
// _.import('tmp/core/modules/hello.module.js').then(modules => {
//   console.log(modules, modules[TIME]);
// });

import('./modules/hello.module').then(modules => {
  console.log(modules);
});
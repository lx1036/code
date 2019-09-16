// Karma configuration
// Generated on Tue Nov 06 2018 13:49:37 GMT+0800 (China Standard Time)
/**
 * @link https://github.com/monounity/karma-typescript
 *
 *  yarn karma start ./src/app/demo/karma/karma.conf.js --fail-on-empty-test-suite
 */
module.exports = function(config) {
  config.set({

    // base path that will be used to resolve all patterns (eg. files, exclude)
    basePath: '.',

    /**
     * Configuration for karma-typescript
     *
     * https://github.com/monounity/karma-typescript/blob/master/README.md#advanced-configuration
     */
    karmaTypescriptConfig: {
      tsconfig: './tsconfig.json',
    },

    // frameworks to use
    // available frameworks: https://npmjs.org/browse/keyword/karma-adapter
    frameworks: ['jasmine', "karma-typescript"],

    /**
     * files configuration
     *
     * https://karma-runner.github.io/3.0/config/files.html
     */
    // list of files / patterns to load in the browser
    files: [
      '*.spec.ts',
      // 'substract2.ts',
      'substract/**/*.ts',
    ],


    // list of files to exclude
    exclude: [
    ],


    // preprocess matching files before serving them to the browser
    // available preprocessors: https://npmjs.org/browse/keyword/karma-preprocessor
    preprocessors: {
      '*.ts': ['karma-typescript'],
      'substract/**/*.ts': ['karma-typescript'],
    },

    // typescriptPreprocessor: {
    //   options: {
    //     sourceMap: true, // generate source maps
    //     noResolve: true // enforce type resolution
    //   },
    //   transformPath: function(path) {
    //     return path.replace(/\.ts$/, '.js');
    //   }
    // },

    // test results reporter to use
    // possible values: 'dots', 'progress'
    // available reporters: https://npmjs.org/browse/keyword/karma-reporter
    reporters: ['dots', 'karma-typescript', 'progress'],


    // web server port
    port: 9876,


    // enable / disable colors in the output (reporters and logs)
    colors: true,


    // level of logging
    // possible values: config.LOG_DISABLE || config.LOG_ERROR || config.LOG_WARN || config.LOG_INFO || config.LOG_DEBUG
    logLevel: config.LOG_INFO,


    // enable / disable watching file and executing tests whenever any file changes
    autoWatch: true,


    // start these browsers
    // available browser launchers: https://npmjs.org/browse/keyword/karma-launcher
    browsers: ['ChromeHeadless'],


    // Continuous Integration mode
    // if true, Karma captures browsers, runs the tests and exits
    singleRun: false,

    // Concurrency level
    // how many browser should be started simultaneous
    concurrency: Infinity
  })
};

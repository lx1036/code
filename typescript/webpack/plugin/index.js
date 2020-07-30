/**
 * Plugins API: https://www.webpackjs.com/api/plugins/
 * Write a Plugin: https://www.webpackjs.com/contribute/writing-a-plugin/
 * Demo: https://fengmiaosen.github.io/2017/03/21/webpack-core-code/
 */

const webpack = require('webpack'); //访问 webpack 运行时(runtime)
const configuration = require('./webpack.config.js');

let compiler = webpack(configuration);

const pluginName = 'ConsoleLogOnBuildWebpackPlugin';

class ConsoleLogOnBuildWebpackPlugin {
  apply(compiler) {
    compiler.hooks.run.tap(pluginName, compilation => {
      console.log("webpack 构建过程开始！");
    });
  }
}

// 一个 JavaScript 命名函数。
function MyExampleWebpackPlugin() {

}

// 在插件函数的 prototype 上定义一个 `apply` 方法。
MyExampleWebpackPlugin.prototype.apply = function(compiler) {
  // 指定一个挂载到 webpack 自身的事件钩子。
  compiler.plugin('webpacksEventHook', function(compilation /* 处理 webpack 内部实例的特定数据。*/, callback) {
    console.log("This is an example plugin!!!");

    // 功能完成后调用 webpack 提供的回调。
    callback();
  });
};


function FileListPlugin(options) {}

FileListPlugin.prototype.apply = function(compiler) {
  compiler.plugin('emit', function(compilation, callback) {
    // 在生成文件中，创建一个头部字符串：
    var filelist = 'In this build:\n\n';

    // 遍历所有编译过的资源文件，
    // 对于每个文件名称，都添加一行内容。
    for (var filename in compilation.assets) {
      filelist += ('- '+ filename +'\n');
    }

    // 将这个列表作为一个新的文件资源，插入到 webpack 构建中：
    compilation.assets['filelist.md'] = {
      source: function() {
        return filelist;
      },
      size: function() {
        return filelist.length;
      }
    };

    console.log('test');

    callback();
  });
};





(new ConsoleLogOnBuildWebpackPlugin()).apply(compiler);

FileListPlugin.apply(compiler);


compiler.run(function(err, stats) {
  // console.log(stats, err);
});

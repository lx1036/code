/**
 * ../../node_modules/.bin/webpack-dev-server ./main.js --config ./build/webpack.config.js
 * ../../node_modules/.bin/webpack ./main.js --config ./build/webpack.config.js
 */
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const CleanWebpackPlugin = require('clean-webpack-plugin');
const webpack = require('webpack');
const VueLoaderPlugin = require('vue-loader/lib/plugin');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = {
  mode: 'development',
  devtool: 'inline-source-map',
  devServer: {
    contentBase: path.resolve(__dirname, 'dist'),
    hot: true,
    port: 3001
  },
  entry: {
    app: 'index.js',
  },
  output: {
    filename: '[name].[hash:8].bundle.js',
    path: path.resolve(__dirname, 'dist'),
  },
  resolve: {
    alias: {
      vue$: 'vue/dist/vue.runtime.esm.js'
    },
  },
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        use: [
          'cache-loader', // cache-loader 用于缓存loader编译的结果
          'thread-loader', // thread-loader 使用 worker 池来运行loader，每个 worker 都是一个 node.js 进程
          'babel-loader'
        ]
      },
      {
        test: /\.(scss|sass|css)$/,
        use: [
          'style-loader', // 将 JS 字符串生成为 style 节点
          'css-loader', // 将 CSS 转化成 CommonJS 模块
          {
            loader: 'sass-loader', // 将 Sass 编译成 CSS，默认使用 Node Sass
            options: {
              implementation: require("sass")
            }
          },
          'postcss-loader'
        ]
      },
      {
        test: /\.(jpe?g|png|gif)$/i,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 4096,
              fallback: {
                loader: 'file-loader',
                options: {
                  name: 'img/[name].[hash:8].[ext]'
                }
              }
            }
          }
        ]
      },
      {
        test: /\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 4096,
              fallback: {
                loader: 'file-loader',
                options: {
                  name: 'media/[name].[hash:8].[ext]'
                }
              }
            }
          }
        ]
      },
      {
        test: /\.(woff2?|eot|ttf|otf)(\?.*)?$/i,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 4096,
              fallback: {
                loader: 'file-loader',
                options: {
                  name: 'fonts/[name].[hash:8].[ext]'
                }
              }
            }
          }
        ]
      },
      {
        test: /\.vue$/,
        use: [
          'cache-loader',
          'thread-loader',
          {
            loader: 'vue-loader', // vue-loader 用于解析.vue文件
            options: {
              compilerOptions: {
                preserveWhiteSpace: false
              }
            }
          }
        ]
      }
    ]
  },
  plugins: [
    new CleanWebpackPlugin([path.resolve(__dirname, 'dist')], {root: path.resolve(__dirname, '.')}),
    new webpack.DefinePlugin({
      'process.env': {
        VUE_APP_BASE_URL: JSON.stringify('http://localhost:3000'),
        NODE_ENV: JSON.stringify('development')
      }
    }),
    new HtmlWebpackPlugin({
      title: '搭建Vue开发环境',
      template: path.resolve(__dirname, './index.html')
    }),
    new webpack.HotModuleReplacementPlugin(),
    new VueLoaderPlugin(),
    // new BundleAnalyzerPlugin({analyzerMode: 'static'}), // 该插件生成依赖报告
  ]
};


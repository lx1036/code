

/**
 *  ../../../../node_modules/.bin/webpack-dev-server --config ./webpack.config.js --watch
 */


const path = require('path');

module.exports = {
  entry: './observable.ts',
  devtool: 'inline-source-map',
  mode: 'development',
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: 'ts-loader',
        exclude: /node_modules/
      }
    ]
  },
  resolve: {
    extensions: ['.ts', '.js']
  },
  output: {
    filename: 'main.js',
    path: path.resolve(__dirname, 'dist')
  }
};
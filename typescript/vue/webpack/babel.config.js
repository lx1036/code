module.exports = {
  presets: [
    '@babel/preset-env'
  ],
  plugins: [
    '@babel/plugin-syntax-dynamic-import' // 打包模块懒加载
  ]
};

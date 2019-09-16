// cd /Users/liuxiang/rightcapital/rightcapital/router/vue/demo/defineProperty
// ../../../node_modules/.bin/ts-node index.ts

import {Watcher} from "./watcher";

/**
 * @see https://github.com/answershuto/learnVue/blob/master/docs/%E5%93%8D%E5%BA%94%E5%BC%8F%E5%8E%9F%E7%90%86.MarkDown
 */



function defineReactive (obj, key, val, cb) {
  const dep = new Dep();

  Object.defineProperty(obj, key, {
    enumerable: true,
    configurable: true,
    get: function reactiveGetter() {
      /*....依赖收集等....*/
      /*Github:https://github.com/answershuto*/
      dep.addSub(Dep.target);
      return val
    },
    set: function reactiveSetter(newVal) {
      if (val === newVal) return;
      // val = newVal;
      // cb(newVal);/*订阅者收到消息的回调*/
      dep.notify();
    }
  })
}

function observe(value, cb) {
  Object.keys(value).forEach((key) => defineReactive(value, key, value[key] , cb));
}

class Vue {
  public _data: any;

  constructor(options) {
    this._data = options.data;
    observe(this._data, options.render);
    new Watcher();
  }
}

let app = new Vue({
  el: '#app',
  data: {
    text: 'text',
    text2: 'text2'
  },
  render(){
    console.log('视图更新啦～');
    // console.log(this.text, this.text2);
  }
});

app._data.text = 'text3';

class Dep {
  public subs: Array<any>;
  public static target: any;

  constructor () {
    /* 用来存放Watcher对象的数组 */
    this.subs = [];
  }

  /* 在subs中添加一个Watcher对象 */
  addSub (sub) {
    this.subs.push(sub);
  }

  /* 通知所有Watcher对象更新视图 */
  notify () {
    this.subs.forEach((sub) => {
      sub.update();
    })
  }
}



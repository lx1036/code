/**
 * ./node_modules/.bin/karma start vue/vue3/karma.conf.js --single-run
 * yarn jest -c ./vue/vue3/jest.config.js
 * @see https://github.com/zzz945/write-vue3-from-scratch/blob/master/doc/zh-cn.md
 */
import {VNode} from "./vnode";
import {Watcher} from "./watcher";
import {clearTarget, createProxy, setTarget} from "./proxy";
import {Dep} from "./dep";

export class Vue3 {
  constructor(options) {
    this.$options = options;
    this.initProps();
    // this.proxy = this.initDataProxy();
    this.proxy = createProxy(this);
    this.initWatcher();
    this.initWatch();

    return this.proxy;
  }

  /**
   * @see https://stackoverflow.com/questions/37714787/can-i-extend-proxy-with-an-es2015-class
   */
  initDataProxy() {
    const data = this.$data = this.$options.data ? this.$options.data(): {};
    const props = this.$props;
    const computed = this.$options.computed || {};
    
    const createDataProxyHandler = (path) => {
      return {
        set: (data, key, value) => {
          const fullPath = path ? path + '.' + key : key;
          
          const prev = data[key];
          data[key] = value;
  
          if (prev !== value) {
            this.notify(fullPath, prev, value);
          }
          
          return true;
        },
        get: (data, key) => {
          const fullPath = path ? path + '.' + key : key;
          
          this.collect(fullPath);
          
          if (!!data[key] && typeof data[key] === 'object') {
            return new Proxy(data[key], createDataProxyHandler(fullPath));
          } else {
            return data[key];
          }
        },
        deleteProperty: (target, key) => {
          if (key in target) {
            const fullPath = path ? path + '.' + key : key;
            const pre = target[key];
            delete target[key];
      
            this.notify(fullPath, pre);
          }
    
          return true;
        }
      };
    };
    
    return new Proxy(this, {
      set: (target, key, value, receiver) => {
        if (key in props) { // check in props firstly
          return createDataProxyHandler().set(props, key, value);
        } else if (key in data) { // check in data secondly
          return createDataProxyHandler().set(data, key, value);
        } else {
          this[key] = value;
        }

        return true;
      },
      get: (target, key, receiver) => {
        const methods = this.$options.methods || {};

        if (key in props) {
          return createDataProxyHandler().get(props, key);
        } else if (key in data) { // 收集模板中用了 data 的属性到依赖集合中
          return createDataProxyHandler().get(data, key);
        } else if (key in computed) {
          return computed[key].call(this.proxy);
        } else if (key in methods) {
          return methods[key].bind(this.proxy);
        }

        return this[key];
      },
    });
  }
  
  initProps() {
    this.$props = {};
    const {props, propsData} = this.$options;
    
    if (!props) {
      return;
    }
    
    props.forEach((prop) => {
      this.$props[prop] = propsData[prop];
    });
  }
  
  collect(key) {
    this.collected = this.collected || {};
    
    if (!this.collected[key]) {
      this.$watch(key, this.update.bind(this)); // 依赖收集
      this.collected[key] = true;
    }
    
    /*if (this.$target) {
      this.$watch(key, this.$target.update.bind(this.$target));
    }*/
  }

  initWatcher() {
    this.deps = {};
  }
  
  initWatch() {
    const watch = this.$options.watch || {};
    const data = this.$data;
    const computed = this.$options.computed || {};
    
    for (let key in watch) {
      const handler = watch[key];
      
      if (key in data) {
        this.$watch(key, handler.bind(this.proxy));
      } else if (key in computed) {
        new Watcher(this.proxy, computed[key], handler);
      } else {
        throw 'the watching key must be keys of data or computed';
      }
    }
    
  }

  notify(key, prev, current) {
    (this.dataNotifyChain[key] || []).forEach((cb) => cb(prev, current));
  }

  $watch(key, cb) {
    if (!this.deps[key]) {
      this.deps[key] = new Dep();
    }
    
    this.deps[key].addSub(new Watcher(this.proxy, key, cb));
    
    // this.dataNotifyChain[key] = this.dataNotifyChain[key] || [];
    // this.dataNotifyChain[key].push(cb);
  }

  $mount(root) {
    this.$el = root;
    // const vnode = render.call(this.proxy, this.createElement.bind(this));
    // this.$el = this.createDOMElement(vnode);
    //
    // if (root) {
    //   root.appendChild(this.$el);
    // }
    setTarget(this);
    this.update(); // first-time render and trigger 'mounted' hook
    clearTarget();
    
    // mounted lifecycle
    const {mounted} = this.$options;
    mounted && mounted.call(this.proxy);

    return this;
  }

  createElement(tagName, attributes, children) {
    const components = this.$options.components || {};
    
    if (tagName in components) {
      return new VNode(tagName, attributes, children, components[tagName]);
    }
    
    return new VNode(tagName, attributes, children);
  }

  createDOMElement(vnode) {
    if (vnode.componentOptions) {
      const componentInstance = new Vue3(Object.assign({}, vnode.componentOptions,{propsData: vnode.attributes.props}));
      vnode.componentInstance = componentInstance;
      componentInstance.$events = (vnode.attributes || {}).on || {};
      componentInstance.$mount();
      
      return componentInstance.$el;
    }
    
    const element = document.createElement(vnode.tagName);
    element.__vue__ = this;

    // Element attributes
    for (let key in vnode.attributes) {
      element.setAttribute(key, vnode.attributes[key]);
    }

    // Set DOM event listener
    const events = (vnode.attributes || {}).on || {};
    for (let key in events) {
      element.addEventListener(key, events[key]);
    }
    

    if (!Array.isArray(vnode.children)) {
      element.textContent = vnode.children + '';
    } else {
      vnode.children.forEach((child) => {
        if (typeof child === 'string') {
          element.textContent = child;
        } else {
          element.appendChild(this.createDOMElement(child));
        }
      });
    }

    return element;
  }
  
  $emit(...options) {
    const [event, ...payload] = options;
    const cb = this.$events[event];
    
    if (cb) {
      cb(...payload);
    }
  }

  update() {
    const parent = (this.$el || {}).parentElement;
    
    if (this.$options.render) {
      const vnode = this.$options.render.call(this.proxy, this.createElement.bind(this));
      const oldElement = this.$el;
      this.$el = this.patch(null, vnode);
  
      if (parent) {
        parent.replaceChild(this.$el, oldElement);
      }
    }
  }

  patch(oldVnode, newVnode) {
    return this.createDOMElement(newVnode);
  }
}

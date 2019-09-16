let _target = null;

export function setTarget(target) {
  _target = target;
}

export function clearTarget() {
  _target = null;
}
export function createProxy(vue) {
  const collect = (key) => {
    if (_target) {
      vue.$watch(key, _target.update.bind($target));
    }
  };
  
  const createDataProxyHandler = (path) => {
    return {
      set: (data, key, value) => {
        const fullPath = path ? path + '.' + key : key;
        
        const prev = data[key];
        data[key] = value;
        
        if (prev !== value) {
          vue.notify(fullPath, prev, value);
        }
        
        return true;
      },
      get: (data, key) => {
        const fullPath = path ? path + '.' + key : key;
  
        collect(fullPath);
        
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
  
  const handler = {
    set: (target, key, value) => {
      if (key in props) { // check in props firstly
        return createDataProxyHandler().set(props, key, value);
      } else if (key in data) { // check in data secondly
        return createDataProxyHandler().set(data, key, value);
      } else {
        this[key] = value;
      }
    
      return true;
    },
    get: (target, key) => {
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
  };
  
  return new Proxy(this, handler);
}

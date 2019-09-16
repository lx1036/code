
export class Watcher {
  constructor(vm, computedHandler, watchHandler) {
    this.vm = vm;
    this.computedHandler = computedHandler;
    this.watchHandler = watchHandler;
    
    vm.$target = this;
    this.value = this.computedHandler.call(vm);
    vm.$target = null;
  }
  
  update() {
    const old = this.value;
    this.value = this.computedHandler.call(this.vm);
    this.watchHandler.call(this.vm, old, this.value);
  }
}

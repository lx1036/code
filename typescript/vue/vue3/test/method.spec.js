import {Vue3} from "../src";


describe('Vue methods', function () {
  it('basic usage', function () {
    const vm = new Vue3({
      methods: {
        hello(vm) {
          return {self: this, msg: 'hello', name: vm.name};
        }
      },
      data() {
        return {
          name: 'world'
        }
      }
    });

    let hello = vm.hello(vm);
    expect(hello.self).toEqual(vm);
    expect(hello.msg).toEqual('hello');
    expect(hello.name).toEqual('world');
  });
});

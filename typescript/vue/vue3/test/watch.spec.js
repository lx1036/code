import {Vue3} from "../src";

describe('watch', () => {
  let cb;
  
  beforeEach(() => {
    cb = jasmine.createSpy('cb');
  });
  
  it('watch data', () => {
    const vm = new Vue3({
      data() {
        return {a: 0};
      },
      watch: {
        a(pre, current) {
          cb(pre, current);
        }
      }
    });
    
    vm.a = 1;
    expect(cb).toHaveBeenCalledWith(0, 1);
  });
  
  it('watch computed', () => {
    const vm = new Vue3({
      data() {
        return {a: 0};
      },
      computed: {
        b() {
          return this.a + 1;
        },
      },
      watch: {
        b(pre, current) {
          cb(pre, current);
        }
      }
    });
  
    expect(vm.b).toEqual(1);
    vm.a = 1;
    expect(vm.b).toEqual(2);
    expect(cb).toHaveBeenCalledWith(1, 2);
  });
});

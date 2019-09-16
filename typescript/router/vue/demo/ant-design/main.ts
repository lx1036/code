import Vue from 'vue';
import App from './App.vue';

Vue.config.productionTip = false;


let data = {name: 'test'};
let obj = {foo: 'bar'};
let freeze = Object.freeze(obj);

let vm = new Vue({
  data: data,
  // data: freeze,
  created: function() {
    console.log('created: ' + this.name);
  },
  computed: {
    reversedName: function (): string {
      return this.name.split('').reverse().join('');
    }
  },
  render: (h) => h(App),
}).$mount('#app');

console.log(vm.name);

data.name = 'test2';
console.log(vm.name, vm.reversedName);

vm.name = 'test3';
console.log(vm.name, data.name, vm.reversedName);

vm.b = 'b';
console.log(vm.b);

// obj.foo = 'bar2';
// console.log(vm.foo);

vm.$watch('name', (newValue, oldValue) => {
  setTimeout(() => {
    console.log('newValue: ' + newValue, 'oldValue: ' + oldValue);
  }, 1000);
});
vm.name = 'test4';

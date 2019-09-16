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

import Vue from 'vue'
import App from './views/App.vue'
import router from './router/router'
import store from './vuex/store';

new Vue({
  render: h => h(App),
  router,
  store,
}).$mount('#app');

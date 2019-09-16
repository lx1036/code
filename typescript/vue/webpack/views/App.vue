<template>
  <div class="App">
    <div>
      <router-link to="/Home">go Home</router-link>
      <router-link to="/About">go About</router-link>
    </div>
    <div>
      <p>{{getCounter}}</p>
      <button @click="add">add</button>
    </div>
    <div>
      <router-view></router-view>
    </div>
  </div>
</template>
<script>
  import { mapActions, mapGetters } from 'vuex'
  export default {
    name: 'App',
    data() {
      return {};
    },
    computed: {
      ...mapGetters(['getCounter'])
    },
    methods: {
      ...mapActions(['add'])
    }
  };
</script>
<style lang="scss" scoped>
  .App {
    text-align: center;
    color: skyblue;
    font-size: 28px;
  }
</style>

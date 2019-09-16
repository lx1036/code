<template>
  <div id="app">
    <p>asdfasdfsdaf</p>
<!--    <resize-observer @element-resize="handleResize"></resize-observer>-->
    <recycle-scroller>
      <template slot="before">
        Before
      </template>
      <template slot="after">
        After
      </template>
    </recycle-scroller>
  </div>
</template>

<script lang="ts">
  import { Component, Prop, Vue } from 'vue-property-decorator';
  import ResizeObserver from './components/resize-observer.vue';
  import RecycleScroller from "./components/recycle-scroller.vue";

  @Component({
    components: {
      'resize-observer': ResizeObserver,
      'recycle-scroller': RecycleScroller,
    }
  })
  export default class App extends Vue {
    public beforeCreate() {
      console.log('beforeCreate');
    }

    public created() {
      console.log('created');
    }

    public beforeMount() {
      console.log('beforeMount');
    }

    public mounted() {
      console.log('mounted');
    }

    public beforeDestroy() {
      console.log('beforeDestroy');
    }

    public destroyed() {
      console.log('destroy');
    }

    public handleResize(layout: any) {
      console.log('resized', layout.width, layout.height);
    }
  }
</script>

<style lang="scss">
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}
#nav {
  padding: 30px;
  a {
    font-weight: bold;
    color: #2c3e50;
    &.router-link-exact-active {
      color: #42b983;
    }
  }
}
</style>

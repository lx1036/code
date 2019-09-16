<template>
  <div class="home">
<!--    <img alt="Vue logo" src="../assets/logo.png">-->
<!--    <HelloWorld msg="Welcome to Your Vue.js + TypeScript App"/>-->
    <Counter></Counter>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import HelloWorld from '@/components/HelloWorld.vue'; // @ is an alias to /src
import Counter from '@/components/Counter.vue';

@Component({
  components: {
    HelloWorld,
    Counter,
  },
})
export default class Home extends Vue {}
</script>

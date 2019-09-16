<template>
  <div class="vue-recycle-scroller">
    <div v-if="$slots.before" class="vue-recycle-scroller-slot">
      <slot name="before"></slot>
    </div>

    <div class="vue-recycle-scroller-item-wrapper">
      <div v-for="view of pool" class="vue-recycle-scroller-item-view">
        <slot></slot>
      </div>
    </div>

    <div v-if="$slots.after" class="vue-recycle-scroller-slot">
      <slot name="after"></slot>
    </div>

    <resize-observer @notify="handleResize"></resize-observer>
  </div>
</template>

<script lang="ts">
  /**
   * @see Slots: https://cn.vuejs.org/v2/guide/components-slots.html
   * @see Slots 实现原理：https://ustbhuangyi.github.io/vue-analysis/extend/slot.html
   */
  import { Component, Prop, Vue } from 'vue-property-decorator';
  import ResizeObserver from './resize-observer.vue';

  @Component({
    name: 'recycle-scroller',
    components: {'resize-observer': ResizeObserver},
    directives: {},
  })
  export default class RecycleScroller extends Vue {
    public pool = [];
    public totalSize = 0;
    private ready = false;

    handleResize() {
      this.$emit('resize');

      if (this.ready) {
        this.updateVisibleItems(false);
      }
    }

    updateVisibleItems(checkItem: boolean) {

    }
  }
</script>

<style scoped>

</style>

<template>
  <div class="resize-observer" tabindex="-1"></div>
</template>

<script lang="ts">
  import { Component, Prop, Vue } from 'vue-property-decorator';

  @Component
  export default class ResizeObserver extends Vue {
    public name: string = 'resize-observer';

    private elWidth: number = 0;
    private elHeight: number = 0;

    private element: HTMLObjectElement = new HTMLObjectElement();

    public mounted() {
      console.log('ResizeObserver mounted');

      this.$nextTick(() => {
        this.elWidth = (this.$el as HTMLElement).offsetWidth; // border+padding+scrollbar+width
        this.elHeight = (this.$el as HTMLElement).offsetHeight; // border+padding+scrollbar+height

        console.log(this.elWidth, this.elHeight);
      });

      this.element = document.createElement('object');
      this.element.setAttribute('tabindex', '-1');
      this.element.setAttribute('aria-hidden', 'true');
      this.element.type = 'text/html';
      this.element.data = 'about:blank';
      this.element.onload = () => {
        ((this.element.contentDocument as Document).defaultView as Window).addEventListener('resize', this.notify);
      };

      this.$el.appendChild(this.element);
    }

    public beforeDestroy() {
      if (this.element && this.element.onload) {
        ((this.element.contentDocument as Document).defaultView as Window).removeEventListener('resize', this.notify);
      }
    }

    private notify() {
      if ((this.$el as HTMLElement).offsetWidth !== this.elWidth || (this.$el as HTMLElement).offsetHeight !== this.elHeight) {
        this.elWidth = (this.$el as HTMLElement).offsetWidth;
        this.elHeight = (this.$el as HTMLElement).offsetHeight;

        this.$emit('element-resize', {width: this.elWidth, height: this.elHeight});
      }
    }
  }


  /*export default {
    name: "resize-observer"
  }*/
</script>

<style scoped lang="scss">
  .resize-observer {
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;

    border: none;

    background: green;

    position: absolute;
    z-index: -1;

    display: block;
    overflow: hidden;
    opacity: 0;
  }
  
  .resize-observer >>> object {
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;

    position: absolute;
    z-index: -1;

    display: block;
    overflow: hidden;
  }
</style>

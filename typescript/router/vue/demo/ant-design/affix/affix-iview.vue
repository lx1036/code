<template>
<!--  <div ref="placeholderNode" >-->
<!--    <div ref="fixedNode">-->

<!--    </div>-->
<!--  </div>-->


<!--  iview-->
  <div>
    <div ref="point" :class="classes" :style="styles">
      <slot></slot>
    </div>
    <div v-show="slot" :style="slotStyle"></div>
  </div>
</template>

<script lang="ts">
  import { Component, Prop, Vue } from 'vue-property-decorator';
  import {VueConstructor} from 'vue';

  // @Component({
  //   name: 'a-affix',
  // })
  // class AffixComponent extends Vue {
  //   public name = "affix";
  //   private events = ['resize', 'scroll', 'touchstart', 'touchmove', 'touchend', 'pageshow', 'load'];
  //   private eventHandlers = {};
  //   private affixStyle;
  //   private placeholderStyle;
  //   /*public static install = (vue: VueConstructor): void => {
  //     vue.component(AffixComponent.name, AffixComponent)
  //   };*/
  //
  //   @Prop(Number) readonly offsetTop!: number;
  //   @Prop(Number) readonly offset!: number;
  //   @Prop(Number) readonly offsetBottom!: number;
  //   @Prop() readonly target!: Function;
  //   @Prop({default: 'ant-affix'}) readonly prefixCls!: string;
  //
  // }

  /*AffixComponent.install = function (vue: VueConstructor) {
    vue.component(AffixComponent.name, AffixComponent)
  };*/

  // export default AffixComponent;
  const prefixCls = 'ivu-affix';
  /**
   * @link https://github.com/iview/iview/blob/3.x/src/components/affix/affix.vue
   */
  export default {
    name: 'Affix',
    props: {
      offsetTop: {
        type: Number,
        default: 0
      },
      offsetBottom: {
        type: Number
      }
    },
    data: function() {
      return {
        affix: false,
        styles: {},
        slot: false,
        slotStyle: {}
      };
    },
    computed: {
      offsetType () {
        let type = 'top';
        if (this.offsetBottom >= 0) {
          type = 'bottom';
        }
        return type;
      },
      classes () {
        return [
          {
            [`${prefixCls}`]: this.affix
          }
        ];
      }
    },
    mounted(): void {
      window.addEventListener('scroll', this.handleScroll, false);
      window.addEventListener('resize', this.handleScroll, false);
    },
    beforeDestroy(): void {
    },
    methods: {
      handleScroll() {
        const affix = this.affix;
        const scrollTop = getScroll(window, true);
        const elOffset = getOffset(this.$el);
      },
    }
  }

</script>

<style scoped>

</style>

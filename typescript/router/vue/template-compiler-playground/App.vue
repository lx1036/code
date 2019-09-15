<template>
  <div id="app" class="app">
    <section class="editor-wrap editable-wrap">
      <editor @change="handleCodeChange"></editor>
    </section>

    <section class="editor-wrap uneditable-wrap">
      <div>
        <select v-model="displayType">
          <option v-for="(value, key) of types" :key="key" :value="key">{{value}}</option>
        </select>
      </div>
<!--      <div :class="" v-for="(type, key) of types" v-show="displayType === key" :key="key">-->
      <div v-if="displayType">
        <editor :title="types[displayType]" :read-only="true" :ref="`${displayType}Editor`"></editor>
      </div>
    </section>
  </div>
</template>

<script>
  import Editor from './Editor.vue';
  import {compile} from 'vue-template-compiler';

  const types = {
    ast: '抽象语法树(AST)',
    render: '渲染函数(render)',
    staticRenderFns: '静态渲染函数(staticRenderFns)',
    errors: '编译错误信息(errors)'
  };

  export default {
    name: "App",
    data() {
      return {
        displayType: 'ast',
        compiledResult: '',
        types,
      };
    },
    watch: {
      displayType: 'setEditorValue'
    },
    components: {
      'editor': Editor,
    },
    methods: {
      handleCodeChange(code) {
        this.compiledResult = compile(code, {preserveWhitespace: true});
        this.setEditorValue(this.displayType);
      },
      setEditorValue(type) {
        this.$nextTick(() => {
          switch (type) {
            case 'ast':
            default:
              this.$refs.astEditor[0].setValue(this.formatJson(this.compiledResult.ast));
              break;
            case 'render':
              break;
            case 'staticRenderFns':
              break;
            case 'errors':
              break;
          }
        });
      },
      formatJson(json) {
        if (!json) return;
        let cache = [];

        return JSON.stringify(json, (key, value) => {
          if (typeof value === 'object' && value !== null) {
            if (cache.indexOf(value) !== -1) {
              return '[循环引用]';
            }

            cache.push(value);
          }

          return value;
        });
      }
    }
  }
</script>

<style scoped lang="scss">
  .app {
    display: flex;
    height: 100%;
  }

  .editable-wrap {
    margin-right: 50px;
  }

  .editor-wrap {
    flex: 1;
    flex-shrink: 0;
    width: 50%;
  }

  .uneditable-wrap {
    display: flex;
    flex-direction: column;
  }

  .child-editor {
    flex: 1;
  }

  .change-type-button {
    margin: 10px 0;
  }
</style>

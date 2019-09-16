<template>
  <div class="editor-box">
    <div v-if="!readOnly" class="change-theme">
      <span class="title">{{title}}</span>
      <span>切换主题：</span>
      <select v-model="currentTheme">
        <option v-for="(theme, index) of themes" :key="index">{{theme}}</option>
      </select>
      <span>切换语言：</span>
      <select v-model="currentMode">
        <option v-for="(mode, index) of modes" :key="index">{{mode}}</option>
      </select>
    </div>
    <textarea ref="editor"></textarea>

<!--    <div>-->
<!--      <label>-->
<!--      </label>-->
<!--    </div>-->
  </div>
</template>

<script>
  import CodeMirror from 'codemirror';
  // import 'codemirror/lib/codemirror.css';
  import 'codemirror/mode/vue/vue';
  // import 'codemirror/theme/3024-night.css';
  import {debounce} from 'lodash';

  const themes = [
    '3024-day',
    '3024-night',
    'abcdef',
    'ambiance-mobile',
    'ambiance',
  ];

  const modes = [
    'vue',
    'javascript',
  ];

  export default {
    name: "Editor",
    data(vm) {
      return {
        currentMode: vm.mode,
        currentTheme: vm.theme,
        themes,
        modes,
      }
    },
    props: {
      readOnly: {
        type: Boolean,
        default: false,
      },
      mode: {
        type: String,
        default: 'vue'
      },
      theme: {
        type: String,
        default: '3024-day',
      },
      title: {
        type: String,
        default: ''
      }
    },
    methods: {
      setValue(code) {
        this.editor.setValue(code);
      },
      loadTheme(theme) {
        // import(`codemirror/theme/${theme}.css`);
      },
      loadMode(mode) {
        // import(`codemirror/mode/${mode}/${mode}`);
      }
    },
    mounted() {
      // this.loadTheme(this.currentTheme);
      // await this.loadMode(this.currentMode);

      this.editor = CodeMirror.fromTextArea(this.$refs.editor, {
        value: '',
        mode: this.currentMode,
        theme: this.currentTheme,
        lineNumbers: true,
        autofocus: !this.readOnly,
        readOnly: this.readOnly,
        lineWrapping: true,
      });

      this.editor.on('change', debounce(() => {
        this.$emit('change', this.editor.getValue());
      }, 3000, {trailing: true}));
    }
  }
</script>

<style scoped lang="scss">

  .CodeMirror {
    flex: 1;
  }

  .editor-box {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .title {
    font-weight: bold;
    color: green;
  }

  .change-theme {
    margin: 10px 0;
  }
</style>

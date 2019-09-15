import {Vue3} from "../src";


describe('demo', () => {
  it('stage1', () => {
    const vm = new Vue3({
      data() {
        return {a: 0};
      },
      render(h) {
        return h('button', {class: 'btn', on: {'click': this.handleClick}}, this.a);
      },
      methods: {
        handleClick() {
          this.a++;
        }
      }
    }).$mount(document.body);
    
    const button = document.body.querySelector('.btn');
    expect(button.tagName).toEqual('BUTTON');
    button.click();
    expect(vm.$el.textContent).toEqual('1');
    expect(document.body.querySelector('.btn').textContent).toEqual('1');
  
    document.body.removeChild(vm.$el);
  });
  
  it('stage2', () => {
    const vm = new Vue3({
      data() {
        return {a: [{}]};
      },
      render(h) {
        return h('div', {class: 'parent'}, this.a.map((item, index) => {
          return h('div', {}, [
            h('button', {class: 'set-number', on: {'click': () => this.setNumber(item)}}, 'Set number'),
            h('button', {class: 'delete-number',on: {'click': () => this.deleteNumber(item)}}, 'Delete number'),
            h('span', {class: 'number'}, item.number),
            h('button', {class: 'append-row', on: {'click': () => this.appendRow}}, 'Append row'),
            h('button', {class: 'remove-row', on: {'click': () => this.removeRow(index)}}, 'Remove row'),
            h('br', {}, ''),
          ]);
        }));
      },
      methods: {
        setNumber(item) {
          item.number = 0;
        },
        deleteNumber(item) {
          delete item.number;
        },
        appendRow() {
          this.a.push({});
        },
        removeRow(index) {
          this.a.splice(index, 1);
        }
      }
    }).$mount(document.body);
    
    // Assert set/delete span number
    let button = document.body.querySelector('.set-number');
    const span = document.body.querySelector('.number');
    expect(span.textContent).toEqual('undefined');
    button.click();
    // expect(span.textContent).toEqual('0');
    button = document.body.querySelector('.delete-number');
    button.click();
    expect(span.textContent).toEqual('undefined');
    
    // Assert append/remove row
    button = document.body.querySelector('.append-row');
    const div = document.body.querySelector('.parent');
    expect(div.children.length).toEqual(1);
    button.click();
    // expect(div.children.length).toEqual(2);
    button = document.body.querySelector('.remove-row');
    button.click();
    expect(div.children.length).toEqual(1);
  });
});

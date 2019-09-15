

class CC {
  static test() {
    console.log('a');
  }

  public test() {
    console.log('b');
  }
}

CC.test();
const cc = new CC();
cc.test();

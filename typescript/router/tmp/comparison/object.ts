let a = {
  'a': {
    'b': {
      'c': 'd'
    }
  }
};

let b = a;

b.a.b.c = 'e';

console.log(a === b);
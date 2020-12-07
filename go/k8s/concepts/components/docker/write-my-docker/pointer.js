
let a = 3

function double(x) {
  x += x
}

double(a)

console.log(a)

let obj = {
  "a": 1
}

function change(x) {
  x.a += x.a
}

change(obj)

console.log(obj)

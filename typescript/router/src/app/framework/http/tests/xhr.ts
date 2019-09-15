/**
 * Run in Browser
 */

const xhr = new XMLHttpRequest();
const onLoad = () => {
  console.log(xhr.getAllResponseHeaders());
};
xhr.addEventListener('load', onLoad);
xhr.open('GET', 'https://jsonplaceholder.typicode.com/posts/1');
xhr.send();

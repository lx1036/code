





let form = document.getElementById('testform');
form.addEventListener('reset', ($event) => {
  console.log($event);
});

form.addEventListener('submit', ($event) => {
  console.log($event);
});
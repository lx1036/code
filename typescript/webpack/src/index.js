import _ from 'lodash';
import './style.css';
import './style2.scss';
import Icon from './icon.png';
import Data from './data.xml';
import CSVData from './data.csv';

import printMe from './print.js';
import {cube} from './math';


function component() {
  let element = document.createElement('div');
  
  /*element.innerHTML = _.join(['Hello', 'webpack'], ' ');
  element.classList.add('hello');
  
  let icon = new Image();
  icon.src = Icon;
  element.appendChild(icon);
  
  console.log(Data, CSVData);
  
  let btn = document.createElement('button');
  btn.innerHTML = '点击这里，然后查看 console!';
  btn.onclick = printMe;
  
  element.appendChild(btn);*/
  
  let pre = document.createElement('pre');
  pre.innerHTML = ['Hello webpack', '5 cubed is equal to' + cube(5)].join('\n\n');
  element.appendChild(pre);
  
  return element;
}

document.body.appendChild(component());

if (module.hot) {
  module.hot.accept('./print.js', function () {
    console.log('Accepting the updated printMe module!');
    printMe();
  })
} 

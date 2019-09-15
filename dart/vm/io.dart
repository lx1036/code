import 'dart:async';

import 'dart:io';

void main() {
  Timer(Duration(seconds: 1), () => print('timer'));
  print("end of main");

  File myfile = File('myfile.txt');
  myfile
      .rename('yourfile.txt')
      .then((_) => {print('name changed'), print('sadfasdf')});
}

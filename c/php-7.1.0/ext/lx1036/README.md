```shell script
./configure --with-lx1036 --without-iconv
make
./sapi/cli/php -f ext/lx1036/lx1036.php
./sapi/cli/php -r "lx1036_hello_world();"
```

dnl $Id$
dnl config.m4 for extension lx1036

dnl Comments in this file start with the string 'dnl'.
dnl Remove where necessary. This file will not work
dnl without editing.

dnl If your extension references something external, use with:

PHP_ARG_WITH(lx1036, for lx1036 support,
dnl Make sure that the comment is aligned:
[  --with-lx1036             Include lx1036 support])

dnl Otherwise use enable:

dnl PHP_ARG_ENABLE(lx1036, whether to enable lx1036 support,
dnl Make sure that the comment is aligned:
dnl [  --enable-lx1036           Enable lx1036 support])

if test "$PHP_LX1036" != "no"; then
  dnl Write more examples of tests here...

  dnl # --with-lx1036 -> check with-path
  dnl SEARCH_PATH="/usr/local /usr"     # you might want to change this
  dnl SEARCH_FOR="/include/lx1036.h"  # you most likely want to change this
  dnl if test -r $PHP_LX1036/$SEARCH_FOR; then # path given as parameter
  dnl   LX1036_DIR=$PHP_LX1036
  dnl else # search default path list
  dnl   AC_MSG_CHECKING([for lx1036 files in default path])
  dnl   for i in $SEARCH_PATH ; do
  dnl     if test -r $i/$SEARCH_FOR; then
  dnl       LX1036_DIR=$i
  dnl       AC_MSG_RESULT(found in $i)
  dnl     fi
  dnl   done
  dnl fi
  dnl
  dnl if test -z "$LX1036_DIR"; then
  dnl   AC_MSG_RESULT([not found])
  dnl   AC_MSG_ERROR([Please reinstall the lx1036 distribution])
  dnl fi

  dnl # --with-lx1036 -> add include path
  dnl PHP_ADD_INCLUDE($LX1036_DIR/include)

  dnl # --with-lx1036 -> check for lib and symbol presence
  dnl LIBNAME=lx1036 # you may want to change this
  dnl LIBSYMBOL=lx1036 # you most likely want to change this 

  dnl PHP_CHECK_LIBRARY($LIBNAME,$LIBSYMBOL,
  dnl [
  dnl   PHP_ADD_LIBRARY_WITH_PATH($LIBNAME, $LX1036_DIR/$PHP_LIBDIR, LX1036_SHARED_LIBADD)
  dnl   AC_DEFINE(HAVE_LX1036LIB,1,[ ])
  dnl ],[
  dnl   AC_MSG_ERROR([wrong lx1036 lib version or lib not found])
  dnl ],[
  dnl   -L$LX1036_DIR/$PHP_LIBDIR -lm
  dnl ])
  dnl
  dnl PHP_SUBST(LX1036_SHARED_LIBADD)

  PHP_NEW_EXTENSION(lx1036, lx1036.c, $ext_shared,, -DZEND_ENABLE_STATIC_TSRMLS_CACHE=1)
fi

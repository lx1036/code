<?php

error_reporting(-1);

$arr = ['a' => 'a'];
$brr = ['a' => 'a'];

unset($arr['b'], $brr['b']);

var_dump($arr, $brr);

/**
 * https://secure.php.net/manual/zh/function.set-exception-handler.php
 * @param Throwable $exception
 */
function exception_handler(Throwable $exception) {
    echo "Uncaught exception: " , $exception->getMessage(), "\n";
}

/**
 * https://secure.php.net/manual/zh/function.set-error-handler.php
 */
function error_handler($level, $message, $file = '', $line = 0) {
    if (error_reporting() & $level) {
        /**
         * https://secure.php.net/manual/en/errorfunc.constants.php
         */
        var_dump(error_reporting(), $level, $message);
        throw new ErrorException($message, 0, $level, $file, $line);
    }
}

set_exception_handler('exception_handler');
set_error_handler('error_handler');
//throw new Exception('Uncaught Exception');

$c = $a + 1;

try {
    $c = $a + 1;
} catch (Exception $exception) {
//    throw new Exception();
}


var_dump(1);
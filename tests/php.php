<?php

$arr = ['a' => 'a'];
$brr = ['a' => 'a'];

unset($arr['b'], $brr['b']);

var_dump($arr, $brr);


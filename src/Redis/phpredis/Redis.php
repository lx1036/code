<?php

$redis = new Redis();
$redis->connect('127.0.0.1', 6379);
//$redis->del('h');
$redis->hSet('h', 'a', 'x');
$redis->hSet('h', 'b', 'y');
$redis->hSet('h', 'c', 'z');
$redis->hSet('h', 'd', 't');
var_dump($redis->hGetAll('h'));

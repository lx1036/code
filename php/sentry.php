<?php

include __DIR__ . "/vendor/autoload.php";



Sentry\init(['dsn' => 'http://593e3a809c044ddaa9a3c2a1a41de751@localhost:9001/3']);

$e = new Exception('test');
Sentry\captureException($e);

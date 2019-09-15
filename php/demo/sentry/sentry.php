<?php

include __DIR__ . "/../../vendor/autoload.php";

Sentry\init(['dsn' => 'https://438764c986184899ac61d8a21b0929d7@sentry.io/1454874']);


function throwException() {
    throw new InvalidArgumentException('This is an exception asdfdsf.');
}

try {
    throwException();
} catch (Exception $exception) {
    Sentry\captureException($exception);
}


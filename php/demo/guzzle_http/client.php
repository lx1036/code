<?php

include __DIR__ . '/../../vendor/autoload.php';

$content = [
    'code' => 1,
    'message' => [
        'test' => 'test1'
    ],
];
$response = \Symfony\Component\HttpFoundation\Response::create(\GuzzleHttp\json_encode($content), 200, ['content-type' => 'application/json']);


$response->send();

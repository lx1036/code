<?php

include __DIR__ . "/../../vendor/autoload.php";


$client = new GuzzleHttp\Client([
    \GuzzleHttp\RequestOptions::PROXY  => '127.0.0.1:8888',
    \GuzzleHttp\RequestOptions::VERIFY => false,
]);

try {
    $response = $client->get('https://auth.capitect.com/v1/token', [
        \GuzzleHttp\RequestOptions::HEADERS => [
            'Content-Type' => 'application/x-www-form-urlencoded',
        ],
        \GuzzleHttp\RequestOptions::AUTH => [
            'apc01xONd6c2S2N',
            '1472c25de751e485a9c853105391f5206d203616',
            'basic',
        ],
        \GuzzleHttp\RequestOptions::FORM_PARAMS => [
            'userCapid' => 'usr01QkDIVJNhZP',
            'userApiKey' => '8c70f89f2d2838a20917700dfaae9a7639330252',
        ],
    ]);
} catch (Exception $exception) {
    dump($exception->getMessage());die();
}


dump($response->getBody(), $response->getHeaders(), $response->getStatusCode());

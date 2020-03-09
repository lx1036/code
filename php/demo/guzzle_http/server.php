<?php

include __DIR__ . '/../../vendor/autoload.php';

$client = new \GuzzleHttp\Client();
$response = $client->get("localhost:4444");
// __toString(), not getContents()
var_dump($response->getBody()->getContents(),$response->getBody()->__toString(), json_decode($response->getBody()->__toString(),true));


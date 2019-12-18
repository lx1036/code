<?php

include __DIR__ . "/../../../vendor/autoload.php";

use Next\Prometheus\Client\CollectorRegistry;
use Next\Prometheus\Client\Exception\MetricsRegistrationException;
use Next\Prometheus\Client\RenderTextFormat;
use Next\Prometheus\Client\Storage\Apcu;
use Next\Prometheus\Client\Storage\InMemory;
use Next\Prometheus\Client\Storage\Redis;

//dump($_GET);
//$adapter = $_REQUEST["adapter"];

//$request = \Illuminate\Http\Request::createFromGlobals();
//$adapter = $request->get('adapter');
////var_dump($adapter);
//$tmp = "adf";
////dump($adapter);die();
//if ($adapter === 'redis') {
//    Redis::setDefaultOptions(['host' => $_SERVER['REDIS_HOST'] ?? '127.0.0.1']);
//    $tmp = new Redis();
//} elseif ($adapter === 'apcu') {
//    $tmp = new Apcu();
//    var_dump("afdsf");
//} elseif ($adapter === 'in-memory') {
//    $tmp = new InMemory();
//}
//$tmp = null;
$tmp = new Apcu();


//var_dump($tmp);
$registry = new CollectorRegistry($tmp);


try {
    $counter = $registry->registerCounter('test', 'some_counter', 'it increases', ['type']);
} catch (MetricsRegistrationException $e) {
    dump($e->getMessage());
}
$count = 10;
$counter->incBy($count, ['blue']);

try {
    $gauge = $registry->registerGauge('test', 'some_gauge', 'it sets', ['type']);
} catch (MetricsRegistrationException $e) {
    dump($e->getMessage());
}
$gauge->set($count, ['blue']);

try {
    $histogram = $registry->registerHistogram('test', 'some_histogram', 'it observes', ['type'], [0.1, 1, 2, 3.5, 4, 5, 6, 7, 8, 9]);
} catch (MetricsRegistrationException $e) {
    dump($e->getMessage());
}
$histogram->observe(3, ['blue']);
$histogram->observe(1, ['blue']);

$renderer = new RenderTextFormat();
$result = $renderer->render($registry->collect());
header('Content-type: ' . RenderTextFormat::MIME_TYPE);
echo $result;

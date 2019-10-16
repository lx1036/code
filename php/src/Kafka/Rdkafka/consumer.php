<?php

use RdKafka\Consumer;

$consumer = new Consumer();
$consumer->addBrokers('localhost:2181');
$topic = $consumer->newTopic('test2');
$topic->consumeStart(0, RD_KAFKA_OFFSET_BEGINNING);

while (true) {
    $msg = $topic->consume(0, 1000);
    if (null === $msg) {
        continue;
    } elseif ($msg->err) {
        echo $msg->errstr(), "\n";
        break;
    } else {
        echo $msg->payload, "\n";
    }
}

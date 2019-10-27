<?php
$conf = new RdKafka\Conf();

$conf->setDrMsgCb(function ($kafka, $message) {
    file_put_contents("./dr_cb.log", var_export($message, true).PHP_EOL, FILE_APPEND);
});
$conf->setErrorCb(function ($kafka, $err, $reason) {
    file_put_contents("./err_cb.log", sprintf("Kafka error: %s (reason: %s)", rd_kafka_err2str($err), $reason).PHP_EOL, FILE_APPEND);
});

$conf->set('log_level', LOG_DEBUG);
$conf->set('debug', 'all');
$kafka_producer = new RdKafka\Producer($conf);
$kafka_producer->addBrokers('localhost');
$topic = $kafka_producer->newTopic('laravel');
//$topic->produce(RD_KAFKA_PARTITION_UA, 0, 'new payload');

for ($i = 0; $i < 10; $i++) {
    $topic->produce(RD_KAFKA_PARTITION_UA, 0, "Message $i");
}

<?php
# Includes the autoloader for libraries installed with composer
require __DIR__ . '/vendor/autoload.php';


putenv("GOOGLE_APPLICATION_CREDENTIALS=" . __DIR__ . '/e454f8cbd8ec.env.json');
# Imports the Google Cloud client library
use Google\Cloud\Translate\TranslateClient;

# Your Google Cloud Platform project ID
$projectId = 'e454f8cbd8eca70d5e1ef6f0f65f509e8a11bb4e';

# Instantiates a client
$translate = new TranslateClient([
    'projectId' => $projectId
]);

# The text to translate
//$text = file_get_contents(__DIR__ . '/go-module.md');
# The target language

$handle = fopen(__DIR__ . '/go-module-en.md', 'r');


while (!feof($handle)) {
    $bytes = fread($handle, 4000);
    if (!empty($bytes)) {
        # Translates some text into Russian
        $translation = $translate->translate($bytes, [
            'target' => 'zh'
        ]);

        file_put_contents(__DIR__ . '/text.test.md', $translation['text'] . PHP_EOL);
    }
}

fclose($handle);

//while (($line = fgets($handle)) != false) {
//    if (!empty($line)) {
//
//    }
//
//}

//echo 'Text: ' . $text . '
//Translation: ' . $translation['text'];

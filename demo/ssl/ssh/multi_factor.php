<?php

require __DIR__ . '/../../../vendor/autoload.php';

use phpseclib\Net\SSH2;

$ssh = new SSH2('192.168.1.254');
if (!$ssh->login('root', 'password')) {
    exit('Login failed');
}
// this does the same thing as the above
//if (!$ssh->login($username, 'pass1') && !$ssh->login('username', 'code1')) {
//    exit('Login failed');
//}

echo $ssh->exec('pwd');
echo $ssh->exec('ls -la');


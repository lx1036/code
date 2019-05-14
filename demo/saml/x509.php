<?php

include __DIR__ . '/../../vendor/autoload.php';

use \phpseclib\Crypt\RSA;
use \phpseclib\File\X509;



$rsa = new RSA();
$keys = $rsa->createKey();
//dump($keys['privatekey']);


$x509 = new X509();
$crts = $x509->loadX509(file_get_contents(__DIR__ . '/../../resources/keys/commonwealth/commonwealth.prd.crt'));
dump($crts);

$privKey = new RSA();
$privKey->loadKey('-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDMswfEpAgnUDWA74zZw5XcPsWh1ly1Vk99tsqwoFDkLF7jvXy1
dDLHYfuquvfxCgcp8k/4fQhx4ubR8bbGgEq9B05YRnViK0R0iBB5Ui4IaxWYYhKE
8xqAEH2fL+/7nsqqNFKkEN9KeFwc7WbMY49U2adlMrpBdRjk1DqIEW3QTwIDAQAB
AoGBAJ+83cT/1DUJjJcPWLTeweVbPtJp+3Ku5d1OdaGbmURVs764scbP5Ihe2AuF
V9LLZoe/RdS9jYeB72nJ3D3PA4JVYYgqMOnJ8nlUMNQ+p0yGl5TqQk6EKLI8MbX5
kQEazNqFXsiWVQXubAd5wjtb6g0n0KD3zoT/pWLES7dtUFexAkEA89h5+vbIIl2P
H/NnkPie2NWYDZ1YiMGHFYxPDwsd9KCZMSbrLwAhPg9bPgqIeVNfpwxrzeksS6D9
P98tJt335QJBANbnCe+LhDSrkpHMy9aOG2IdbLGG63MSRUCPz8v2gKPq3kYXDxq6
Y1iqF8N5g0k5iirHD2qlWV5Q+nuGvFTafCMCQQC1wQiC0IkyXEw/Q31RqI82Dlcs
5rhEDwQyQof3LZEhcsdcxKaOPOmKSYX4A3/f9w4YBIEiVQfoQ1Ig1qfgDZklAkAT
TQDJcOBY0qgBTEFqbazr7PScJR/0X8m0eLYS/XqkPi3kYaHLpr3RcsVbmwg9hVtx
aBtsWpliLSex/HHhtRW9AkBGcq67zKmEpJ9kXcYLEjJii3flFS+Ct/rNm+Hhm1l7
4vca9v/F2hGVJuHIMJ8mguwYlNYzh2NqoIDJTtgOkBmt
-----END RSA PRIVATE KEY-----');
$pubKey = new RSA();
$pubKey->loadKey($privKey->getPublicKey());
$pubKey->setPublicKey();
// Subject as certificate, including public key
$subject = new X509();
$subject->setDNProp('id-at-organizationName', 'phpseclib demo cert');
$subject->setPublicKey($pubKey);
// Issuer sign Subject using private key
$issuer = new X509();
$issuer->setPrivateKey($privKey);
$issuer->setDN($subject->getDN());
$x509 = new X509();
$x509->setEndDate('lifetime');
$result = $x509->sign($issuer, $subject);
$cert = $x509->saveX509($result);
//dump($cert);
$cert = $x509->loadX509($cert);
//dump($cert);


/*$privKey = new RSA();
extract($privKey->createKey());
$privKey->loadKey($privatekey);

$pubKey = new RSA();
$pubKey->loadKey($publickey);
$pubKey->setPublicKey();

$issuer->setPrivateKey($privKey);
$subject->setPublicKey($pubKey);



$pubKey = new \phpseclib\Crypt\RSA();
$subject = new \phpseclib\File\X509();
$subject->setPublicKey($pubKey); // $pubKey is Crypt_RSA object
$subject->setDN('/O=phpseclib demo cert');

$privKey=new \phpseclib\Crypt\RSA();
$issuer = new \phpseclib\File\X509();
$issuer->setPrivateKey($privKey); // $privKey is Crypt_RSA object
$issuer->setDN('/O=phpseclib demo cert');

$x509 = new \phpseclib\File\X509();
$result = $x509->sign($issuer, $subject);
echo $x509->saveX509($result);*/

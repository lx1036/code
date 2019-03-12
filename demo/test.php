<?php


class AClass {
    public $a = 1;
}


$aclass = new AClass();
$aclass->a = 2;

function test(AClass $aclass) {
    $aclass->a = 3;
}

var_dump($aclass->a);

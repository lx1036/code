<?php

include __DIR__ . "/../vendor/autoload.php";




use Illuminate\Database\Capsule\Manager as Capsule;


$capsule = new Capsule;

$capsule->addConnection([
    'driver'    => 'mysql',
    'host'      => 'localhost',
    'database'  => 'test',
    'username'  => 'root',
    'password'  => '',
    'charset'   => 'utf8',
    'collation' => 'utf8_unicode_ci',
    'prefix'    => '',
]);
// Set the event dispatcher used by Eloquent models... (optional)
//use Illuminate\Events\Dispatcher;
use Illuminate\Container\Container;
//$capsule->setEventDispatcher(new Dispatcher(new Container));

// Make this Capsule instance available globally via static methods... (optional)
$capsule->setAsGlobal();

// Setup the Eloquent ORM... (optional; unless you've used setEventDispatcher())
$capsule->bootEloquent();

Capsule::schema()->create('users', function (\Illuminate\Database\Schema\Blueprint $table) {
    $table->increments('id');
    $table->string('email')->unique();
    $table->timestamps();
});





<?php


return [

    'connections' => [
        'mysql' => [ // connection name
            'driver' => 'mysql', // driver
            'host' => 'localhost',
            'port' => 3306,
            'database' => 'database',
            'username' => 'username',
            'password' => 'password',
            'charset' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'engine' => 'innodb',
        ],
        'mysql::read' => [ // connection name
            'driver' => 'mysql', // driver
            'host' => 'localhost',
            'port' => 3306,
            'database' => 'database_read',
            'username' => 'username',
            'password' => 'password',
            'charset' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'engine' => 'innodb',
        ],
    ]

];
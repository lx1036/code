<?php

return [
    'default' => [
        'driver' => 'session'
    ],
    'drivers' => [
        // name
        'session' => [
            'driver' => 'session',
            'provider' => 'database',
        ],

        'token' => [
            'driver' => 'token'
        ]
    ],
    'providers' => [
        'database' => [
            'driver' => 'database',

        ]
    ]
];

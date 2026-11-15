<?php

return [

    /*
    |--------------------------------------------------------------------------
    | Cross-Origin Resource Sharing (CORS) Configuration
    |--------------------------------------------------------------------------
    |
    | Here you may configure your settings for cross-origin resource sharing
    | or "CORS". This determines which domains are allowed to access your
    | application's resources from a different domain. You may adjust
    | these settings as needed.
    |
    | To learn more: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
    |
    */

    'paths'                    => ['*'],

    'allowed_methods'          => ['*'],

    'allowed_origins'          => [
        'http://localhost:3000',
        'http://localhost:5173',
        'http://news.localhost:5173',
        'http://dashboard.localhost:5173',
        'https://perimatrik.org',
        'https://news.perimatrik.org',
        'https://dashboard.perimatrik.org',
        'https://dashboard.perimatrik.org/',
        'https://beta.perimatrik.org'
    ],

    'allowed_origins_patterns' => [
        '#^https?://.*\.perimatrik\.org$#',
        '#^https?://.*\.localhost:5173$#'
    ],

    'allowed_headers'          => ['*'],

    'exposed_headers'          => [],

    'max_age'                  => 0,

    'supports_credentials'     => true,

];

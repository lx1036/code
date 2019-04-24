<?php

/**
 * @see https://stripe.com/docs/api?lang=php
 */

include __DIR__ . "/../../vendor/autoload.php";


\Stripe\Stripe::setApiKey('sk_test_mH9mRWjy6GIPQtVPXaVdYbbd00njycYrHu');


$customers = \Stripe\Customer::all();
dump($customers->values());

$accounts = \Stripe\Account::all(["limit" => 3]);
//dump($accounts->values());

$account = \Stripe\Account::create([
    "type" => "custom",
    "country" => "US",
    "email" => "bob@example.com",
    "requested_capabilities" => ["platform_payments"]
]);
dump($account->__toArray(true));



/*

\Stripe\Stripe::setClientId('ca_ErKlnwyWVbQx66AhlT59uQlOIohmzN5M');


if (isset($_GET['code'])) {
    // The user was redirected back from the OAuth form with an authorization code.
    $code = $_GET['code'];

    try {
        $resp = \Stripe\OAuth::token([
            'grant_type' => 'authorization_code',
            'code' => $code,
        ]);
    } catch (\Stripe\Error\OAuth\OAuthBase $e) {
        exit("Error: " . $e->getMessage());
    }

    $accountId = $resp->stripe_user_id;

    echo "<p>Success! Account <code>$accountId</code> is connected.</p>\n";
    echo "<p>Click <a href=\"?deauth=$accountId\">here</a> to disconnect the account.</p>\n";

} elseif (isset($_GET['error'])) {
    // The user was redirect back from the OAuth form with an error.
    $error = $_GET['error'];
    $error_description = $_GET['error_description'];

    echo "<p>Error: code=" . htmlspecialchars($error, ENT_QUOTES) . ", description=" . htmlspecialchars($error_description, ENT_QUOTES) . "</p>\n";
    echo "<p>Click <a href=\"?\">here</a> to restart the OAuth flow.</p>\n";

} elseif (isset($_GET['deauth'])) {
    // Deauthorization request
    $accountId = $_GET['deauth'];

    try {
        \Stripe\OAuth::deauthorize([
            'stripe_user_id' => $accountId,
        ]);
    } catch (\Stripe\Error\OAuth\OAuthBase $e) {
        exit("Error: " . $e->getMessage());
    }

    echo "<p>Success! Account <code>" . htmlspecialchars($accountId, ENT_QUOTES) . "</code> is disconnected.</p>\n";
    echo "<p>Click <a href=\"?\">here</a> to restart the OAuth flow.</p>\n";

} else {
    $url = \Stripe\OAuth::authorizeUrl([
        'scope' => 'read_only',
    ]);
    echo "<a href=\"$url\">Connect with Stripe</a>\n";
}
*/

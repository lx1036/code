<?php

include __DIR__ . "/../vendor/autoload.php";


/* PDO
try {
    $pdo = new PDO('mysql:host=localhost;port=3306;dbname=open_api', 'testing', 'testing');
} catch (PDOException $exception) {
    echo "Can't connect database=open_api";
}

try {

    $select_query = 'select `users`.`name`, `users`.`password` from `users` where `users`.`email` = :email';
    $insert_query = 'insert into `users` (`users`.`name`, `users`.`email`, `users`.`password`, `users`.`remember_token`) values (:name, :email, :password, :remember_token)';
    $update_query = 'update `users` set `users`.`name` = :name, `users`.`password` = :password where `users`.`email` = :email';
    $update_query_placeholders = 'update `users` set `users`.`name` = ?, `users`.`password` = ? where `users`.`email` = ?';
    $delete_query = '';

    $select_bindings = [
        ':email' => 'test@test.com',
    ];
    $bindings = [
        ':name' => 'name1',
        ':email' => 'email2@email.com',
        ':password' => 'password1',
        ':remember_token' => 'abc',
    ];
    $update_bindings = [
        ':name' => 'name1_updated',
        ':email' => 'email1@email.com',
        ':password' => 'password_updated',
    ];

    $update_bindings_placeholders = [
        'name_updated_placeholder',
        'password_updated',
        'email1@email.com',
    ];

//    $pdo_statement = $pdo->prepare($insert_query);
//    $pdo_statement = $pdo->prepare($update_query);
//    $pdo_statement = $pdo->prepare($update_query_placeholders);
    $pdo_statement = $pdo->prepare($select_query);

    $pdo_statement->setFetchMode(PDO::FETCH_ASSOC);

    foreach ($select_bindings as $key => $binding) {
        $pdo_statement->bindValue(is_string($key) ? $key : $key + 1, $binding, is_int($binding) ? PDO::PARAM_INT : PDO::PARAM_STR);
    }

    if ($pdo_statement->execute()) {
        $results = $pdo_statement->fetchAll();
    }

    $count = $pdo_statement->rowCount();
} catch (Exception $exception) {
    var_dump($exception->getMessage());
}


if ($count > 0) {
    echo $count . ' rows is affected.';
} else {
    echo '0 row is affected.';
}

if ($results) {
    var_dump($results);
}

*/

$app = new \Next\Foundation\Container\Application();
$database_manager = new \Next\Database\DatabaseManager($app);
$connection = $database_manager->connection();

$query = 'select `users`.`name`, `users`.`password` from `users` where `users`.`email` = :email';

/** @var array $rows */
$rows = $connection->select($query, [
    ':email' => 'test@test.com'
]);

var_dump($rows);

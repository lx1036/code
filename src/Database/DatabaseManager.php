<?php


namespace Next\Database;


use Next\Database\Orm\Query\Connection\Connection;
use Next\Database\Orm\Query\Connection\ConnectionInterface;
use Next\Database\Orm\Query\Connection\MySqlConnection;
use Next\Foundation\Container\Application;

class DatabaseManager
{
    /** @var Connection[] */
    protected $connections;

    protected $app;

    public function __construct(Application $app)
    {
        $this->app = $app;
    }

    public function connection($name = null): Connection
    {
        $name = $name ?? $this->getDefaultConnectionName();

        $this->connections[$name] = $this->makeConnection($name);

        return $this->connections[$name];
    }


    protected function makeConnection($name): Connection
    {
        $config = $this->configuration($name);

        $pdo = new \PDO($config['host'], $config['username'], $config['password']);

        switch ($config['driver']) {
            case 'mysql':
                return new MySqlConnection($pdo, $config);
        }
    }

    protected function configuration($name): array
    {
        return $this->app['config']['database.connections.' . $name];
    }

    private function getDefaultConnectionName()
    {
        return $this->app['config']['database.default'];
    }
}
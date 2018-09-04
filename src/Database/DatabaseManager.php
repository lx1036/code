<?php


namespace Next\Database;


use Illuminate\Contracts\Support\Arrayable;
use Illuminate\Support\Str;
use Next\Database\Orm\Query\Connection\Connection;
use Next\Database\Orm\Query\Connection\ConnectionInterface;
use Next\Database\Orm\Query\Connection\MySqlConnection;
use Next\Foundation\Container\Application;
use Closure;
use PDO;

class DatabaseManager
{
    /** @var Connection[] */
    protected $connections;

    protected $app;

    protected $custom_creators = [];

    public function __construct(Application $app)
    {
        $this->app = $app;
    }

    public function connection($name = null): ConnectionInterface
    {
        $name = $name ?? $this->getDefaultConnectionName();

        $this->connections[$name] = $this->resolve($name);

        return $this->connections[$name];
    }


    protected function makeConnection($name): Connection
    {
        $config = $this->getConfig($name);

        $pdo = new PDO($config['host'], $config['username'], $config['password']);

        switch ($config['driver']) {
            case 'mysql':
                return new MySqlConnection($pdo, $config);
        }
    }

    protected function getConfig($name): array
    {
        return $this->app['config']['database.connections.' . $name];
    }

    private function getDefaultConnectionName()
    {
        return $this->app['config']['database.default'];
    }

    public function resolve($name): ConnectionInterface
    {
        $config = $this->getConfig($name);

        if (isset($this->customCreators[$config['driver']])) {
            return $this->callCustomCreator($name, $config);
        }

        $method = 'create' . Str::studly($name) . 'Driver';

        if (method_exists($this, $method)) {
            return $this->{$method}($config);
        }

        throw new \InvalidArgumentException('');
    }

    public function driver()
    {

    }

    /**
     *
     *
     * @param array $config
     * @return MySqlConnection
     */
    public function createMysqlDriver(array $config): MySqlConnection
    {
        $pdo = new PDO($config['host'], $config['username'], $config['password']);

        return new MySqlConnection($pdo, $config);
    }

    public function extend($name, Closure $callback)
    {
        $this->custom_creators[$name] = $callback;
    }

    private function callCustomCreator($name, array $config)
    {
        return $this->custom_creators[$config['driver']]($name, $config);
    }
}
<?php


namespace Next\Database\Orm\Query\Connection;


use Next\Database\Orm\Query\Grammar\Grammar;
use Next\Database\Orm\Query\Grammar\GrammarInterface;
use Next\Database\Orm\Query\PostProcessor\Processor;
use Next\Database\Orm\Query\Builder as QueryBuilder;
use PDO;
use PDOStatement;

class Connection implements ConnectionInterface
{
    /** @var \Closure|\PDO */
    protected $pdo;

    /** @var array */
    protected $config;

    protected $grammar;

    protected $post_processor;

    protected $is_query_log = false;

    protected $query_logs;

    protected $fetch_mode = PDO::FETCH_ASSOC;

    protected $records_changed = false;


    /**
     * @param \Closure|\PDO $pdo
     * @param array $config
     */
    public function __construct($pdo, $config = [])
    {
        $this->pdo = $pdo;
        $this->config = $config;
        $this->grammar = new Grammar();
        $this->post_processor = new Processor();
    }

    public function table(string $table)
    {
        $query = new QueryBuilder($this, $this->grammar, $this->post_processor);

        $query->from($table);
    }

    public function select(string $query, array $bindings = []): array
    {

        /**
         * 1. set the fetch mode
         * 2. prepare the bindings
         */

        /** @var \PDOStatement $statement */
        $statement = $this->getPdo()->prepare($query);
        $statement->setFetchMode($this->fetch_mode);

        $this->bindValues($statement, $this->prepareBindings($bindings));

        $statement->execute();

        return $statement->fetchAll();
    }


    protected function bindValues(PDOStatement $statement,array $bindings)
    {
        foreach ($bindings as $binding => $value) {
            $statement->bindValue(is_string($binding) ? $binding : $binding + 1, $value,is_int($value) ? \PDO::PARAM_INT : \PDO::PARAM_STR);
        }
    }

    public function getQueryGrammar(): Grammar
    {
        return $this->grammar;
    }

    /**
     * value is bool or DateTimeInterface
     *
     * @param array $bindings
     * @return array
     */
    protected function prepareBindings(array $bindings): array
    {
        $date_format = $this->getQueryGrammar()->getDateFormat();

        foreach ($bindings as $binding => $value) {
            if (is_bool($value)) {
                $bindings[$binding] = (int) $value;
            } elseif ($value instanceof \DateTimeInterface) {
                $bindings[$binding] = $value->format($date_format);
            }
        }

        return $bindings;
    }

    public function insert(string $query, array $bindings = []): int
    {
        $this->statement($query, $bindings);
    }

    /**
     * @param string $query
     * @param array $bindings
     *
     * @return int
     */
    public function update(string $query, array $bindings = []): int
    {
        $this->statement($query, $bindings);
    }

    public function delete(string $query, array $bindings = []): int
    {
        $this->statement($query, $bindings);
    }

    private function getPdo(): \PDO
    {
        if ($this->pdo instanceof \Closure) {
            $this->pdo = call_user_func($this->pdo);
        }

        return $this->pdo;
    }

    //region log query


    public function getQueryLog()
    {
        return $this->query_logs;
    }

    public function flushQueryLog()
    {
        $this->query_logs = [];
    }

    public function enableQueryLog()
    {
        $this->is_query_log = true;
    }

    public function disableQueryLog()
    {
        $this->is_query_log = false;
    }

    public function logQuery($query, $bindings, $elapsed_time = null)
    {
        if ($this->is_query_log) {
            $this->query_logs[] = compact('query', 'bindings', 'elapsed_time');
        }
    }

    public function statement($query, array $bindings): int
    {








    }

    protected function affectingStatement($query, array $bindings = [])
    {
        /** @var \PDOStatement $statement */
        $statement = $this->getPdo()->prepare($query);

        foreach ($bindings as $binding => $value) {
            $statement->bindValue(is_string($binding) ? $binding : $binding + 1, $value,is_int($value) ? \PDO::PARAM_INT : \PDO::PARAM_STR);
        }

        $statement->execute();

        if (($count = $statement->rowCount()) > 0) {
            if (! $this->records_changed) {
                $this->records_changed = true;
            }
        }

        return $count;
    }

    public function run($query, $bindings, \Closure $callback)
    {
        $start = microtime(true);

        try {
            $result = $callback($query, $bindings);
        } catch (\Exception $exception) {

        }



        $this->logQuery($query, $bindings, $this->getElapsedTime($start));

        return $result;
    }

    protected function getElapsedTime($start): float
    {
        return round((microtime(true) - $start) * 1000, 2);
    }

    //endregion
}
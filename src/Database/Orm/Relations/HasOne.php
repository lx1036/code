<?php


namespace Next\Database\Orm\Relations;


use Next\Database\Model;
use Next\Database\Orm\Builder;

class HasOne extends Relation
{
    protected $foreign_key;

    protected $local_key;

    protected $parent;

    public function __construct(Builder $query, Model $parent, $foreign_key, $local_key)
    {
        parent::__construct($query);

        $this->parent = $parent;
        $this->foreign_key = $foreign_key;
        $this->local_key = $local_key;
    }


}
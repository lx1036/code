<?php


namespace Next\Database\Orm\Relations;


use Next\Database\Orm\Builder;

class BelongsTo extends Relation
{
    public function __construct(Builder $query)
    {
        parent::__construct($query);


    }
}
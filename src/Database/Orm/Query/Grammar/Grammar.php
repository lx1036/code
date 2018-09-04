<?php


namespace Next\Database\Orm\Query\Grammar;


class Grammar implements GrammarInterface
{

    public function getDateFormat()
    {
        return 'Y-m-d H:i:s';
    }
}
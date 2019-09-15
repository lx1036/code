<?php


namespace Next\Algorithms\Tree\RadixTree;


class Node implements \Countable
{
    /** @var \Next\Algorithms\Tree\RadixTree\Node[] $children */
    public $children;

    public $char;

    public $parent;

    /** @var bool $isWord */
    public $isWord;

    public function __construct($char = '', &$parent = null, $isWord = false)
    {
        $this->char = $char;
        $this->parent = &$parent;
        $this->isWord = $isWord;
        $this->children = [];
    }

    public function count()
    {
        // TODO: Implement count() method.
    }
}

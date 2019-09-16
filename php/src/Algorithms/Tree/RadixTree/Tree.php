<?php


namespace Next\Algorithms\Tree\RadixTree;

/**
 * @see https://github.com/SiroDiaz/DataStructures/blob/master/DataStructures/Trees/TrieTree.php
 */
class Tree implements \Countable
{
    /** @var \Next\Algorithms\Tree\RadixTree\Node $root */
    private $root;
    private $numberWords;

    /** @var int $size */
    private $size;

    public function __construct()
    {
        $this->root = new Node();
        $this->numberWords = 0;
        $this->size = 0;
    }


    public function count()
    {
        // TODO: Implement count() method.
    }

    public function add(string $search)
    {
        $search = trim($search);
        $length = mb_strlen($search);

        if ($length == 0) {
            return;
        }

        $current = &$this->root;

        for ($i = 0; $i < $length; $i++) {
            $char = mb_substr($search, $i, 1, 'UTF-8');
            if (!isset($current->children[$char])) {
                if ($i === $length - 1) {
                    $current->children[$char] = new Node($char, $current, true);
                    $this->numberWords++;
                } else {
                    $current->children[$char] = new Node($char, $current, false);
                }

                $this->size++;
            } else {
                if ($i == $length - 1) {
                    if ($current->children[$char]->isWord === false) {
                        $current->children[$char]->isWord = true;
                        $this->numberWords++;
                    }
                }
            }

            $current = &$current->children[$char];
        }
    }

    public function size(): int
    {
        return $this->size;
    }

    public function wordCount(): int
    {
        return $this->numberWords;
    }

    public function contains(string $search): bool
    {
        $search = trim($search);
        $length = mb_strlen($search);

        if ($length == 0) {
            return false;
        }

        /** @var \Next\Algorithms\Tree\RadixTree\Node $current */
        $current = &$this->root;

        for ($i = 0; $i < $length; $i++) {
            $char = mb_substr($search, $i, 1, 'UTF-8');

            if (isset($current->children[$char])) {
                $current = &$current->children[$char];

                if ($i == $length - 1 && $current->isWord) {
                    return  true;
                }
            } else {
                return false;
            }
        }

        return false;
    }

    public function startsWith(string $search): bool
    {
        $search = trim($search);
        $length = mb_strlen($search);

        if ($length == 0) {
            return false;
        }

        /** @var \Next\Algorithms\Tree\RadixTree\Node $current */
        $current = &$this->root;

        for ($i = 0; $i < $length; $i++) {
            $char = mb_substr($search, $i, 1, 'UTF-8');

            if (isset($current->children[$char])) {
                if ($i == $length - 1) {
                    return  true;
                } else {
                    $current = &$current->children[$char];
                }
            } else {
                return false;
            }
        }

        return false;
    }

    public function withPrefix(string $prefix): array
    {
        $node = $this->getNodeFromPrefix($prefix);
        $words = [];

        if ($node !== null) {
            if ($node->isWord) {
                $words[] = $prefix;
            }

            foreach ($node->children as $char => $child) {
                $words = $this->traverseWithPrefix($prefix . $char, $node->children[$char],$words);
            }
        }

        return $words;
        /*$search = trim($search);
        if (empty($search)) {
            return $this->allWord();
        }

        $length = mb_strlen($search);
        $current = &$this->root;
        $result = [];
        
        for ($i = 0; $i < $length; $i++) {
            $char = mb_substr($search, $i, 1, 'UTF-8');
            
            if (isset($current->children[$char])) {
                $current = &$current->children[$char];
                if ($current->isWord) {
                    $result[] = $search;
                    $current = &$current->children[$char];
                }
            } else {
                return [];
            }
        }
        
        
        do {
            if ($current->isWord) {
                $result[] = $search . $current->char;
            }
            foreach ($current->children as $child) {
                
            }
            
            $current = &$current->children[$char];
        } while(empty($current->children));*/
        
    }

    private function traverseWithPrefix($word, Node $node = null, &$words = [])
    {
        if ($node->isWord) {
            $words[] = $word;
        }

        if (empty($node->children)) {
            return $words;
        }

        foreach ($node->children as $char => $child) {
            $words = $this->traverseWithPrefix($word . $char, $node->children[$char], $words);
        }

        return $words;
    }

    private function getNodeFromPrefix($prefix)
    {
        if ($this->size === 0) {
            return null;
        }

        $i = 0;
        $current = $this->root;
        $prefixLength = mb_strlen(trim($prefix));
        while ($i < $prefixLength) {
            $char = mb_substr($prefix, $i, 1, 'UTF-8');
            if (!isset($current->children[$char])) {
                return null;
            }

            $current = $current->children[$char];
            $i++;
        }

        return $current;
    }
}

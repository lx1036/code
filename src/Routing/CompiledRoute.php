<?php


namespace Next\Routing;


class CompiledRoute implements \Serializable
{
    private $variables;
    private $tokens;
    private $staticPrefix;
    private $regex;
    private $pathVariables;
    private $hostVariables;
    private $hostRegex;
    private $hostTokens;

    /**
     * @param string      $staticPrefix  The static prefix of the compiled route
     * @param string      $regex         The regular expression to use to match this route
     * @param array       $tokens        An array of tokens to use to generate URL for this route
     * @param array       $pathVariables An array of path variables
     * @param string|null $hostRegex     Host regex
     * @param array       $hostTokens    Host tokens
     * @param array       $hostVariables An array of host variables
     * @param array       $variables     An array of variables (variables defined in the path and in the host patterns)
     */
    public function __construct(string $staticPrefix, string $regex, array $tokens, array $pathVariables, string $hostRegex = null, array $hostTokens = [], array $hostVariables = [], array $variables = [])
    {
        $this->staticPrefix = $staticPrefix;
        $this->regex = $regex;
        $this->tokens = $tokens;
        $this->pathVariables = $pathVariables;
        $this->hostRegex = $hostRegex;
        $this->hostTokens = $hostTokens;
        $this->hostVariables = $hostVariables;
        $this->variables = $variables;
    }

    public function serialize()
    {
        // TODO: Implement serialize() method.
    }

    public function unserialize($serialized)
    {
        // TODO: Implement unserialize() method.
    }

    public function getHostRegex()
    {
    }

    public function getUriRegex()
    {
    }
}
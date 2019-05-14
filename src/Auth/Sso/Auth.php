<?php


namespace Next\Auth\Sso;


class Auth
{
    public function __construct()
    {
    }

    /**
     * Process the SAML response from the IdP
     *
     * @param string|null $requestId The ID of the AuthNRequest sent by this SP to the IdP
     *
     */
    public function processResponse($requestId = null)
    {

    }
}

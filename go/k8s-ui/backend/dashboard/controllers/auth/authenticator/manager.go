package authenticator

import (
	"github.com/spf13/viper"
	"strings"
)

func AuthenticationModes() []AuthenticationMode {
	modes := strings.Split(viper.GetString("common.authentication-mode"), ",")
	var authModes []AuthenticationMode
	for mode := range modes {
		authModes = append(authModes, AuthenticationMode(mode))
	}

	return authModes
}

func AuthenticationSkippable() bool {
	return viper.GetBool("common.enable-skip-login")
}




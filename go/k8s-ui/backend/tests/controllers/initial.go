package controllers

import ()

const (
	// signed by rsa-private.pem, exp 10 years
	Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhZG1pbiIsImV4cCI6MTg5NjYwODc2NCwiaWF0IjoxNTgxMjQ4NzY0LCJpc3MiOiJrOHMtdWkifQ.IoBWPDJYORTHiRPNOPXNSRsHnOYyJVo8pP0zar0J_Xmx3-wEYTOYvaJxOTCK4ZZB-F0Xbf6owY1DMzDcM1wPqg"
)

func init() {
	/*configFile := "app.conf"
	filename, _ := filepath.Abs("../../")
	viper.SetConfigType("ini")
	file := fmt.Sprintf("%s/conf/%s", filename, configFile)
	viper.SetConfigFile(file)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	initial.InitRsaKey(viper.GetString("default.RsaPrivateKey"), viper.GetString("default.RsaPublicKey"))*/

	initial.InitRsaKey("../../apikey/rsa-private.pem", "../../apikey/rsa-public.pem")
}

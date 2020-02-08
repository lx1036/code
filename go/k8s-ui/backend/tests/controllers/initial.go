package controllers

import (
	_ "k8s-lx1036/k8s-ui/backend/database/lorm"
	"k8s-lx1036/k8s-ui/backend/initial"
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

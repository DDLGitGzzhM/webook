package startup

import "github.com/spf13/viper"

func InitViperDev() {
	viper.SetConfigFile("webook/config/config/dev.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func InitViperRemote() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3",
		"127.0.0.1:12379", "/webook")
	if err != nil {
		return
	}
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

package config

import (
	"github.com/spf13/viper"
)

// Go 中大写开头才能被外部引用

// Load and unmarshal config with given config path
func NewFromConfig(configName string, path string, configStruct interface{}) (*viper.Viper, error) {

	v := viper.New() // 创建新的 Viper 实例

	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath(path)

	// 读取配置文件并装载到 v 实例中
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// unmarshal 到结构体实例
	if err := v.Unmarshal(&configStruct); err != nil {
		return nil, err
	}

	return v, nil
}

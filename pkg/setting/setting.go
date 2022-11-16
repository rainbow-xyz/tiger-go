package setting

import "github.com/spf13/viper"

type Setting struct {
	vp *viper.Viper
}

func NewSetting(confDir string) (*Setting, error) {
	vp := viper.New()
	vp.SetConfigFile(confDir)

	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}
	return &Setting{vp: vp}, nil
}

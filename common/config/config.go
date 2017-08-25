package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func Decode(filename string, conf interface{}) error {

	config_path, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	//config_dir := filepath.Dir(config_path)

	// 打开配置文件
	configFile, err := os.Open(config_path)
	if err != nil {
		return fmt.Errorf("%v %v", config_path, err)
	}
	defer configFile.Close()

	buf := make([]byte, 3)
	if _, err := io.ReadFull(configFile, buf); err != nil {
		return fmt.Errorf("%v %v", config_path, err)
	}
	if bytes.Equal(buf, []byte{0xEF, 0xBB, 0xBF}) == false {
		configFile.Seek(0, 0)
	}

	_, err = toml.DecodeReader(configFile, conf)
	if err != nil {
		return err
	}
	return nil
}

func Eecode(filename string, conf interface{}) error {

	config_path, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	//config_dir := filepath.Dir(config_path)

	configFile, err := os.OpenFile(config_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return fmt.Errorf("%v %v", config_path, err)
	}
	defer configFile.Close()

	e := toml.NewEncoder(configFile)
	return e.Encode(conf)
}

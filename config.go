package mysocks

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// Config 定义配置类型
type Config struct {
	LocalListenAddr  string `json:"LocalListenAddr"`
	RemoteServerAddr string `json:"RemoteServerAddr"`
	Password         string `json:"password"`
}

var configPath string

func init() {
	dir, _ := os.Getwd()
	configFileName := "config.json"
	if len(os.Args) == 2 {
		configFileName = os.Args[1]
	}
	configPath = path.Join(dir, configFileName)
}

// ReadConfigFromFile 读取配置文件中的信息到 Config 变量，json 反序列化
func (config *Config) ReadConfigFromFile() {
	if _, err := os.Stat(configPath); !os.IsNotExist(err) { // 如果 os.IsNotExist() 的返回值为 true,代表文件不存在
		log.Printf("从配置文件 %s 中读取配置信息\n", configPath)
		file, err := os.Open(configPath)
		if err != nil {
			log.Fatalf("打开配置文件 %s 出错\n", configPath)
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(config)
		if err != nil {
			log.Fatalf("配置文件的 json 格式不合法，请检查配置文件：%s\n", configPath) // Fatalf 用来写日志后，用 os.exit(1) 进行退出
		}
	}
}

// WriteConfigToFile 写配置信息到 config.json 文件，json 序列化
func (config *Config) WriteConfigToFile() {
	configData, _ := json.MarshalIndent(config, "", "	") // configData 类型为 []byte
	err := ioutil.WriteFile(configPath, configData, 0666)
	if err != nil {
		log.Fatalf("保存配置信息到文件 %s 失败，失败原因： %s", configPath, err)
	}
	log.Printf("保存配置信息到文件：%s 成功\n", configPath)
}

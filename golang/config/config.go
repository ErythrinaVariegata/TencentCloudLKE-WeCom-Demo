package config

import (
	"flag"
	"fmt"
	"os"
)

// Config 服务全局配置
var Config GlobalConfig

// GlobalConfig 全局配置结构体
type GlobalConfig struct {
	WxToken               string
	WxEncodingAESKey      string
	WxCorpID              string
	WxAppSecret           string
	TencentCloudLKEAppKey string
}

// IsValid 校验配置项是否都有数据
func (c *GlobalConfig) IsValid() bool {
	if c.WxToken == "" || c.WxEncodingAESKey == "" || c.WxCorpID == "" || c.WxAppSecret == "" ||
		c.TencentCloudLKEAppKey == "" {
		return false
	}
	return true
}

func Init() {
	// 定义命令行参数
	flag.StringVar(&Config.WxToken, "wx_token", "", "WeCom App Token")
	flag.StringVar(&Config.WxEncodingAESKey, "wx_encodingaeskey", "", "WeCom App Encoding AES Key")
	flag.StringVar(&Config.WxCorpID, "wx_corpid", "", "WeCom Corp ID")
	flag.StringVar(&Config.WxAppSecret, "wx_appsecret", "", "WeCom App Secret")
	flag.StringVar(&Config.TencentCloudLKEAppKey, "lke_appkey", "", "TencentCloud LKE App Key")

	// 解析命令行参数
	flag.Parse()

	// 如果命令行参数为空，尝试从环境变量获取
	if Config.WxToken == "" {
		Config.WxToken = os.Getenv("WX_TOKEN")
	}
	if Config.WxEncodingAESKey == "" {
		Config.WxEncodingAESKey = os.Getenv("WX_ENCODING_AES_KEY")
	}
	if Config.WxCorpID == "" {
		Config.WxCorpID = os.Getenv("WX_CORP_ID")
	}
	if Config.WxAppSecret == "" {
		Config.WxAppSecret = os.Getenv("WX_APP_SECRET")
	}
	if Config.TencentCloudLKEAppKey == "" {
		Config.TencentCloudLKEAppKey = os.Getenv("TENCENT_CLOUD_LKE_APP_KEY")
	}

	// 验证必要参数是否都已设置
	if !Config.IsValid() {
		fmt.Println("Missing required parameters. Please provide them via command line flags or environment variables:")
		fmt.Println("Environment variables:")
		fmt.Println("  WX_TOKEN")
		fmt.Println("  WX_ENCODING_AES_KEY")
		fmt.Println("  WX_CORP_ID")
		fmt.Println("  WX_APP_SECRET")
		fmt.Println("  TENCENT_CLOUD_LKE_APP_KEY")
		fmt.Println("\nOr command line flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

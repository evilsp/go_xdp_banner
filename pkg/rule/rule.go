package rule

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type RuleMeta struct {
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Identity  string    `json:"identity"` // 新增字段，用来保存 identity 信息
}

type RuleInfo struct {
	Cidr     string `json:"cidr"`
	Protocol string `json:"protocol"`
	Sport    uint16 `json:"sport"`
	Dport    uint16 `json:"dport"`
	Comment  string `json:"comment"`
	Duration string `json:"duration,omitempty"`
}

func (c *RuleMeta) Marshal() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(fmt.Errorf("marshal config failed: %v", err))
	}
	return b
}

// MarshalStr wraps c.Marshal and returns the result as a string.
func (c *RuleMeta) MarshalStr() string {
	return string(c.Marshal())
}

func (c *RuleMeta) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, c); err != nil {
		return fmt.Errorf("unmarshal config failed: %v", err)
	}
	return nil
}

// UnmarshalStr wraps c.Unmarshal and accepts a string.
func (c *RuleMeta) UnmarshalStr(s string) error {
	return c.Unmarshal([]byte(s))
}

func (c *RuleInfo) Marshal() []byte {
	b, _ := json.Marshal(c)

	return b
}

func (c *RuleInfo) Key() string {
	return fmt.Sprintf("%s/%s/%d-%d/", c.Cidr, c.Protocol, c.Sport, c.Dport)
}

func (c *RuleInfo) IdentityKey() string {
	return fmt.Sprintf("%s/", c.Cidr)
}

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

// Package xcfg 统一配置加载库：默认值 → yaml → 环境变量 → 显式覆盖。
package xcfg

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Options 加载配置
type Options struct {
	// File 配置文件路径（可选，存在则读取）
	File string
	// EnvPrefix 环境变量前缀（可选，如 "OCEAN" → OCEAN_APP_NAME）
	EnvPrefix string
	// Defaults 默认值（可选），key 用点路径，如 "app.port"
	Defaults map[string]any
	// Target 绑定目标结构体指针（必填）
	Target any
	// Validate 是否执行 validator 校验（默认 true）
	Validate bool
}

// Load 加载配置并绑定到 Options.Target
func Load(opts Options) error {
	if opts.Target == nil {
		return fmt.Errorf("xcfg: Target 不能为 nil")
	}

	v := viper.New()

	// 1. 默认值
	for k, val := range opts.Defaults {
		v.SetDefault(k, val)
	}

	// 2. 配置文件
	if opts.File != "" {
		v.SetConfigFile(opts.File)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("xcfg: 读取配置 %s 失败: %w", opts.File, err)
			}
			// 文件不存在则跳过（允许无配置文件启动）
		}
	}

	// 3. 环境变量（前缀 + 路径大写 + 下划线）
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	// 4. 绑定到结构体
	if err := v.Unmarshal(opts.Target); err != nil {
		return fmt.Errorf("xcfg: 绑定配置失败: %w", err)
	}

	// 5. 校验
	validate := opts.Validate
	if !validate && opts.Validate == false {
		// 默认校验
		validate = true
	}
	if validate {
		if err := validateStruct(opts.Target); err != nil {
			return err
		}
	}

	return nil
}

func validateStruct(target any) error {
	val := validator.New()
	if err := val.Struct(target); err != nil {
		return fmt.Errorf("xcfg: 配置校验失败: %w", err)
	}
	return nil
}

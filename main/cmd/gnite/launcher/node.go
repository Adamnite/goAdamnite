package launcher

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"github.com/adamnite/go-adamnite/adm"
	"github.com/adamnite/go-adamnite/adm/adamconfig"
	"github.com/adamnite/go-adamnite/internal/utils"
	"github.com/adamnite/go-adamnite/node"
	"github.com/naoina/toml"
	"github.com/urfave/cli/v2"
)

var (
	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

type adamConfig struct {
	Adamnite adamconfig.Config
	Node     node.Config
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Version = "1.0.0"
	return cfg
}

func defaultDemoNodeConfig() node.Config {
	cfg := node.DefaultDemoConfig
	cfg.Version = "demo-1.0.1"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, adamConfig) {
	cfg := adamConfig{
		Adamnite: adamconfig.Defaults,
		Node:     defaultNodeConfig(),
	}

	// Load config file.
	if file := ctx.String(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	utils.SetAdamniteConfig(ctx, stack, &cfg.Adamnite)

	return stack, cfg
}

func makeAdamniteNode(ctx *cli.Context) (*node.Node, adm.AdamniteAPI) {
	stack, cfg := makeConfigNode(ctx)
	adamniteImpl := utils.RegisterAdamniteService(stack, &cfg.Adamnite)
	return stack, adamniteImpl
}

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func loadConfig(file string, cfg *adamConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/angeldswang/recycle"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
)

var (
	thriftPath string
	targetType string
	prettyMode bool
	json       = jsoniter.ConfigCompatibleWithStandardLibrary
)

var rootCmd = &cobra.Command{
	Use:   "recycle",
	Short: "Recycle works like a `thrift store` polishing bytes to original shape",
	Long: `Given the bytes encoded by thrift protocols and the target definition in thrift IDL,
recycle can restore the data with specific type names instead of field numbers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			scanner := bufio.NewScanner(bufio.NewReader(cmd.InOrStdin()))
			for scanner.Scan() {
				args = append(args, scanner.Text())
			}
		}
		shapes, err := recycle.Polish(thriftPath, targetType, args)
		if err != nil {
			return err
		}

		for _, shape := range shapes {
			var out []byte
			if prettyMode {
				// [TODO] https://github.com/json-iterator/go/pull/273
				out, _ = json.MarshalIndent(shape, "", "  ")
			} else {
				out, _ = json.Marshal(shape)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&thriftPath, "thrift", "f", "", "thrift idl file path")
	rootCmd.PersistentFlags().StringVarP(&targetType, "type", "t", "", "target type name in thrift idl")
	rootCmd.PersistentFlags().BoolVarP(&prettyMode, "pretty", "p", false, "use pretty output mode")
}

func initConfig() {
	// TODO
}

func main() {
	Execute()
}

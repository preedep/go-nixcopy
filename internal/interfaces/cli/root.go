package cli

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "nixcopy",
	Short: "Fast file transfer CLI tool",
	Long: `go-nixcopy เป็นเครื่องมือ CLI สำหรับการถ่ายโอนไฟล์ความเร็วสูง
รองรับการถ่ายโอนระหว่าง SFTP, FTPS, Azure Blob Storage และ AWS S3
ด้วยการใช้ streaming เพื่อประหยัดหน่วยความจำ`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

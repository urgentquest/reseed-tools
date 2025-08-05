package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// keygenCmd represents the keygen command
var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate keys for reseed su3 signing and TLS serving",
	Run: func(cmd *cobra.Command, args []string) {
		keygenAction()
	},
}

func init() {
	rootCmd.AddCommand(keygenCmd)

	keygenCmd.PersistentFlags().String("signer", "", "Generate a private key and certificate for the given su3 signing ID (ex. something@mail.i2p)")
	keygenCmd.PersistentFlags().String("tlsHost", "", "Generate a self-signed TLS certificate and private key for the given host")
	viper.BindPFlags(keygenCmd.Flags())
}

// NewKeygenCommand creates a new CLI command for generating keys.

func keygenAction() error {
	signerID := viper.GetString("signer")
	tlsHost := viper.GetString("tlsHost")
	trustProxy := viper.GetBool("trustProxy")

	if signerID == "" && tlsHost == "" {
		fmt.Println("You must specify either --tlsHost or --signer")
		return fmt.Errorf("You must specify either --tlsHost or --signer")
	}

	if signerID != "" {
		if err := createSigningCertificate(signerID); nil != err {
			fmt.Println(err)
			return err
		}
	}

	if trustProxy {
		if tlsHost != "" {
			if err := createTLSCertificate(tlsHost); nil != err {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}

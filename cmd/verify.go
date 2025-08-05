package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"i2pgit.org/idk/reseed-tools/reseed"
	"i2pgit.org/idk/reseed-tools/su3"
)

func I2PHome() string {
	envCheck := os.Getenv("I2P")
	if envCheck != "" {
		return envCheck
	}
	// get the current user home
	usr, err := user.Current()
	if nil != err {
		panic(err)
	}
	sysCheck := filepath.Join(usr.HomeDir, "i2p-config")
	if _, err := os.Stat(sysCheck); nil == err {
		return sysCheck
	}
	usrCheck := filepath.Join(usr.HomeDir, "i2p")
	if _, err := os.Stat(usrCheck); nil == err {
		return usrCheck
	}
	return ""
}

// verifyCmd represents the keygen command
var verifyCmd = &cobra.Command{
	Use:   "verify [flags] filename",
	Short: "Verify a Su3 file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		su3VerifyAction(args[0])
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	keygenCmd.Flags().Bool("extract", false, "Also extract the contents of the su3")
	keygenCmd.Flags().String("signer", getDefaultSigner(), "Your su3 signing ID (ex. something@mail.i2p)")
	keygenCmd.PersistentFlags().String("keystore", filepath.Join(I2PHome(), "/certificates/reseed"), "Path to the keystore")
	viper.BindPFlags(verifyCmd.Flags())
}

func su3VerifyAction(filename string) error {
	su3File := su3.New()

	data, err := os.ReadFile(filename)
	if nil != err {
		return err
	}
	if err := su3File.UnmarshalBinary(data); err != nil {
		return err
	}

	fmt.Println(su3File.String())
	absPath, err := filepath.Abs(viper.GetString("keystore"))
	if nil != err {
		return err
	}
	keyStorePath := filepath.Dir(absPath)
	reseedDir := filepath.Base(absPath)

	// get the reseeder key
	ks := reseed.KeyStore{Path: keyStorePath}

	if viper.GetString("signer") != "" {
		su3File.SignerID = []byte(viper.GetString("signer"))
	}
	log.Println("Using keystore:", absPath, "for purpose", reseedDir, "and", string(su3File.SignerID))

	cert, err := ks.DirReseederCertificate(reseedDir, su3File.SignerID)
	if nil != err {
		fmt.Println(err)
		return err
	}

	if err := su3File.VerifySignature(cert); nil != err {
		return err
	}

	fmt.Printf("Signature is valid for signer '%s'\n", su3File.SignerID)

	if viper.GetBool("extract") {
		// @todo: don't assume zip
		os.WriteFile("extracted.zip", su3File.BodyBytes(), 0o755)
	}
	return nil
}

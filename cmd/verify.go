package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli/v3"
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

func NewSu3VerifyCommand() *cli.Command {
	return &cli.Command{
		Name:        "verify",
		Usage:       "Verify a Su3 file",
		Description: "Verify a Su3 file",
		Action:      su3VerifyAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "extract",
				Usage: "Also extract the contents of the su3",
			},
			&cli.StringFlag{
				Name:  "signer",
				Value: getDefaultSigner(),
				Usage: "Your su3 signing ID (ex. something@mail.i2p)",
			},
			&cli.StringFlag{
				Name:  "keystore",
				Value: filepath.Join(I2PHome(), "/certificates/reseed"),
				Usage: "Path to the keystore",
			},
		},
	}
}

func su3VerifyAction(c *cli.Context) error {
	su3File := su3.New()

	data, err := ioutil.ReadFile(c.Args().Get(0))
	if nil != err {
		return err
	}
	if err := su3File.UnmarshalBinary(data); err != nil {
		return err
	}

	fmt.Println(su3File.String())
	absPath, err := filepath.Abs(c.String("keystore"))
	if nil != err {
		return err
	}
	keyStorePath := filepath.Dir(absPath)
	reseedDir := filepath.Base(absPath)

	// get the reseeder key
	ks := reseed.KeyStore{Path: keyStorePath}

	if c.String("signer") != "" {
		su3File.SignerID = []byte(c.String("signer"))
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

	if c.Bool("extract") {
		// @todo: don't assume zip
		ioutil.WriteFile("extracted.zip", su3File.BodyBytes(), 0755)
	}
	return nil
}

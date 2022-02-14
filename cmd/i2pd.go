//go:build i2pd
// +build i2pd

package cmd

import (
	i2pd "github.com/eyedeekay/go-i2pd/goi2pd"
)

func InitializeI2PD() func() {
	return i2pd.InitI2PSAM(nil)
}

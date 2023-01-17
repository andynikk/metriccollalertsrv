package main

import (
	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
)

func main() {
	arrCert, err := encryption.CreateCert()
	if err != nil {
		constants.Logger.ErrorLog(err)
		return
	}

	encryption.SaveKeyInFile(&arrCert[0], "publicKey.cer")
	encryption.SaveKeyInFile(&arrCert[1], "privateKey.pfx")
}

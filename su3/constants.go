package su3

// SU3 File format constants
// Moved from: su3.go
const (
	minVersionLength = 16

	SigTypeDSA             = uint16(0)
	SigTypeECDSAWithSHA256 = uint16(1)
	SigTypeECDSAWithSHA384 = uint16(2)
	SigTypeECDSAWithSHA512 = uint16(3)
	SigTypeRSAWithSHA256   = uint16(4)
	SigTypeRSAWithSHA384   = uint16(5)
	SigTypeRSAWithSHA512   = uint16(6)

	ContentTypeUnknown   = uint8(0)
	ContentTypeRouter    = uint8(1)
	ContentTypePlugin    = uint8(2)
	ContentTypeReseed    = uint8(3)
	ContentTypeNews      = uint8(4)
	ContentTypeBlocklist = uint8(5)

	FileTypeZIP   = uint8(0)
	FileTypeXML   = uint8(1)
	FileTypeHTML  = uint8(2)
	FileTypeXMLGZ = uint8(3)
	FileTypeTXTGZ = uint8(4)
	FileTypeDMG   = uint8(5)
	FileTypeEXE   = uint8(6)

	magicBytes = "I2Psu3"
)

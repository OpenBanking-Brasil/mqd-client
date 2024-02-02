package crosscutting

import (
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// OFBStruct Defines a base structure for the solution
type OFBStruct struct {
	Pack   string     // Package to be used
	Logger log.Logger // Logger to be used
}

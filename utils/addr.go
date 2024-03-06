package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/unpackdev/solgo/utils/entities"
)

type NamedAddr struct {
	Name  string          `json:"name"`
	Addr  common.Address  `json:"addr"`
	Token *entities.Token `json:"token"`
}

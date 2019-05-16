package controller

import (
	"github.com/aware-hq/azure-key-vault-controller/pkg/controller/azurekeyvaultsecret"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, azurekeyvaultsecret.Add)
}

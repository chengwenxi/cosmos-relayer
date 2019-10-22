package relayer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	mockbank "github.com/cosmos/cosmos-sdk/x/ibc/mock/bank"
)

func ReadPassphraseFromStdin(name string) (string, error) {
	buf := bufio.NewReader(os.Stdin)
	prompt := fmt.Sprintf("Password to sign with '%s':", name)

	passphrase, err := input.GetPassword(prompt, buf)
	if err != nil {
		return passphrase, fmt.Errorf("error reading passphrase: %v", err)
	}

	return passphrase, nil
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()

	ibc.AppModuleBasic{}.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	mockbank.RegisterCdc(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc.Seal()
}

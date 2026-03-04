package crypto

import "github.com/leonklingele/passphrase"

// PassphraseGenerate 生成密码短语
// https://www.eff.org/dice
func PassphraseGenerate(num int, separator ...string) (string, error) {
	if len(separator) == 0 {
		passphrase.Separator = "-"
	} else {
		passphrase.Separator = separator[0]
	}
	return passphrase.Generate(num)
}

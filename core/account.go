package core

type Account struct {
	Address   string
	PublicKey []byte
	SecretKey []byte
	Banlance  uint64
	Nonce     uint64
}

type AccountManager struct {
	Accounts map[string]*Account
}

func NewAccountManager() *AccountManager {
	return &AccountManager{
		Accounts: make(map[string]*Account),
	}
}

func (am *AccountManager) NewAccount() (*Account, error) {
	publicKey, secretKey := "test", "test"
	return &Account{PublicKey: []byte(publicKey), SecretKey: []byte(secretKey)}, nil
}

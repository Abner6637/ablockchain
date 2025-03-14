package core

type Account struct {
	Address   string
	PublicKey []byte
	SecretKey []byte
	Banlance  uint64
	Nonce     uint64
}

type AccountManager struct {
	accounts map[string]*Account
}

func NewAccountManager() *AccountManager {
	return &AccountManager{
		accounts: make(map[string]*Account),
	}
}

func (am *AccountManager) NewAccount() (*Account, error) {
	publicKey, secretKey := "test", "test"
	return &Account{PublicKey: []byte(publicKey), SecretKey: []byte(secretKey)}, nil
}

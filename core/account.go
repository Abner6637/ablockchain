package core

import "log"

type Account struct {
	Address   string
	PublicKey []byte
	SecretKey []byte
	Banlance  uint64
}

type AccountManager struct {
	Accounts map[string]*Account
}

func NewAccountManager() *AccountManager {
	log.Printf("创建新的账户管理器\n")
	return &AccountManager{
		Accounts: make(map[string]*Account),
	}
}

func (am *AccountManager) NewAccount() (*Account, error) {
	publicKey, secretKey := "test", "test"
	log.Printf("创建新账户，publickey: %s\n", publicKey)
	return &Account{PublicKey: []byte(publicKey), SecretKey: []byte(secretKey)}, nil
}

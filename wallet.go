package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
	walletFile         = "wallet.dat"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)

	fullPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

func (w Wallet) MarshalJSON() ([]byte, error) {
	mapStringAny := map[string]any{
		"PrivateKey": map[string]any{
			"D": w.PrivateKey.D.String(),
			"PublicKey": map[string]any{
				"X": w.PrivateKey.X.String(),
				"Y": w.PrivateKey.Y.String(),
			},
		},
		"PublicKey": hex.EncodeToString(w.PublicKey),
	}

	return json.Marshal(mapStringAny)
}

func (w *Wallet) UnmarshalJSON(data []byte) error {
	var temp struct {
		PrivateKey struct {
			D         string `json:"D"`
			PublicKey struct {
				X string `json:"X"`
				Y string `json:"Y"`
			} `json:"PublicKey"`
		} `json:"PrivateKey"`
		PublicKey string `json:"PublicKey"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	d := new(big.Int)
	d.SetString(temp.PrivateKey.D, 10)

	x := new(big.Int)
	x.SetString(temp.PrivateKey.PublicKey.X, 10)

	y := new(big.Int)
	y.SetString(temp.PrivateKey.PublicKey.Y, 10)

	w.PrivateKey = ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		},
	}

	publicKey, err := hex.DecodeString(temp.PublicKey)
	if err != nil {
		return err
	}
	w.PublicKey = publicKey

	return nil
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.X.Bytes(), private.Y.Bytes()...)

	return *private, pubKey
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

package bindings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/sag-enhanced/native-app/src/helper"
)

// a lot of thought into making this process secure
// if you see any security issues, please let me know ASAP
// -> https://sage.party/contact <-

var clientIntent string
var clientSecret string

// trust me, this is required to keep confidentiality in case of id.sage.party compromise
var serverECIntentPublicKey = []byte{48, 89, 48, 19, 6, 7, 42, 134, 72, 206, 61, 2, 1, 6, 8, 42, 134, 72, 206, 61, 3, 1, 7, 3, 66, 0, 4, 217, 237, 95, 239, 38, 161, 158, 33, 241, 194, 181, 97, 179, 197, 202, 84, 167, 104, 239, 1, 199, 72, 252, 31, 50, 139, 245, 233, 28, 141, 138, 135, 95, 135, 252, 168, 38, 54, 127, 229, 60, 245, 77, 211, 192, 133, 193, 86, 170, 177, 100, 113, 34, 51, 32, 151, 208, 81, 28, 28, 213, 245, 103, 180}

const idProtocol = "https"
const idHostname = "id.sage.party"

func (b *Bindings) InitConnect(handover string, resource string) error {
	if identity == nil {
		return errors.New("Identity not loaded")
	}
	if currentUrl == nil || currentUrl.Host == idHostname {
		return fmt.Errorf("this binding must not be called from %s", idHostname)
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		return err
	}

	clientIntent = handover
	clientSecret = base64.RawStdEncoding.EncodeToString(secret)

	query := url.Values{
		"secret":   {clientSecret},
		"handover": {handover},
		"resource": {resource},
	}
	b.ui.Navigate(fmt.Sprintf("%s://%s/#%s", idProtocol, idHostname, query.Encode()))
	return nil
}

func (b *Bindings) ApproveConnect(secret string, approveIntent string, password string) (string, error) {
	// prevent bruteforce and replay attacks
	defer func() {
		clientIntent = ""
		clientSecret = ""
	}()

	if currentUrl == nil || currentUrl.Host != idHostname {
		return "", fmt.Errorf("this binding must be called from %s", idHostname)
	}
	if secret != clientSecret || secret == "" {
		return "", errors.New("Invalid secret")
	}

	intentPK, err := x509.ParsePKIXPublicKey(serverECIntentPublicKey)
	if err != nil {
		return "", err
	}

	parsed, err := jwt.ParseSigned(approveIntent, []jose.SignatureAlgorithm{jose.ES256})
	if err != nil {
		return "", err
	}

	var claims jwt.Claims
	var serverIntent serverIntent
	if err := parsed.Claims(intentPK, &claims, &serverIntent); err != nil {
		return "", err
	}

	// get the thumbprint of the public key
	der := x509.MarshalPKCS1PublicKey(&identity.PrivateKey.PublicKey)
	digest := sha256.Sum256(der)

	// we must make sure that we are approving the correct account
	// the ID isnt that important here since its validated at the start, but why not
	// (I currently dont see a security benefit to this, but it might be useful in the future)
	if err := claims.Validate(jwt.Expected{
		Issuer:      "v4",
		AnyAudience: jwt.Audience{"v4/connect/server-intent"},
		ID:          clientSecret,
		Subject:     base64.RawStdEncoding.EncodeToString(digest[:]),
	}); err != nil {
		return "", err
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	sealed := ""
	if password != "" {
		key := helper.DeriveKey(password, salt)
		aesCipher, err := aes.NewCipher(key)

		if err != nil {
			return "", err
		}

		gcm, err := cipher.NewGCM(aesCipher)
		if err != nil {
			return "", err
		}

		plaintext, err := x509.MarshalPKCS8PrivateKey(identity.PrivateKey)
		if err != nil {
			return "", err
		}

		// now the RSA key is secured with the password
		ciphertext := gcm.Seal(nil, salt[:gcm.NonceSize()], plaintext, nil)

		data := make([]byte, len(salt)+len(ciphertext))
		copy(data, salt)
		copy(data[len(salt):], ciphertext)

		// to prevent a compromise of id.sage.party just immediately decrypting the key
		// using the password it knows, we encrypt the key with the server's public key as well
		// then the only way to decrypt the key is to have the password and the server's private key or
		// to decrypt the key after its been sent to the server
		// however that will require email verification (or server/database compromise ig?)

		serverRSAIntentPublicKey, err := base64.RawStdEncoding.DecodeString(serverIntent.PublicKey)
		if err != nil {
			return "", err
		}
		rsaKey, err := x509.ParsePKCS1PublicKey(serverRSAIntentPublicKey)
		if err != nil {
			return "", err
		}

		sealedData, err := helper.RSASeal(rsaKey, data)
		if err != nil {
			return "", err
		}

		sealed = base64.RawStdEncoding.EncodeToString(sealedData)
	}

	challenge, err := base64.RawStdEncoding.DecodeString(serverIntent.Challenge)
	if err != nil {
		return "", err
	}

	// ID challenge
	if challenge[0] != 0x00 || challenge[1] != 0x02 {
		return "", errors.New("Invalid challenge")
	}

	response, err := identity.Sign(challenge)
	if err != nil {
		return "", err
	}

	return sealed + "." + base64.RawStdEncoding.EncodeToString(response), nil
}

func (b *Bindings) RecoverConnect(secret string, data string, password string) error {
	// we dont use defer to reset the secret, because people can mistype the password
	// trying to mitigate password bruteforce is meaningless here, because that can simply
	// be done somewhere else
	if secret != clientSecret || secret == "" {
		clientSecret = "" // prevent bruteforce and replay attacks
		return errors.New("Invalid secret")
	}

	if currentUrl == nil || currentUrl.Host != idHostname {
		return fmt.Errorf("this binding must be called from %s", idHostname)
	}

	sealed, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	key := helper.DeriveKey(password, sealed[:32])
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return err
	}

	plaintext, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[32:], nil)
	if err != nil {
		return err
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(plaintext)
	if err != nil {
		return err
	}

	identity.PrivateKey = privateKey.(*rsa.PrivateKey)
	return identity.Save(b.fm)
}

type serverIntent struct {
	PublicKey string `json:"spk"`
	Challenge string `json:"challenge"`
}

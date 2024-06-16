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
var idEmail string

// trust me, this is required to keep confidentiality in case of id.sage.party compromise
var serverECIntentPublicKey = []byte{48, 89, 48, 19, 6, 7, 42, 134, 72, 206, 61, 2, 1, 6, 8, 42, 134, 72, 206, 61, 3, 1, 7, 3, 66, 0, 4, 109, 101, 131, 190, 177, 140, 172, 224, 131, 225, 122, 29, 184, 193, 156, 190, 226, 114, 177, 221, 204, 235, 137, 251, 158, 68, 13, 22, 13, 54, 89, 191, 146, 65, 33, 101, 118, 110, 198, 89, 33, 114, 242, 199, 184, 6, 206, 75, 77, 171, 51, 84, 246, 104, 12, 200, 157, 116, 241, 200, 93, 233, 220, 19}
var serverRSAIntentPublicKey = []byte{48, 130, 2, 34, 48, 13, 6, 9, 42, 134, 72, 134, 247, 13, 1, 1, 1, 5, 0, 3, 130, 2, 15, 0, 48, 130, 2, 10, 2, 130, 2, 1, 0, 182, 93, 203, 67, 205, 123, 63, 171, 181, 127, 203, 247, 68, 223, 161, 208, 35, 29, 125, 122, 149, 152, 140, 172, 6, 251, 117, 140, 14, 68, 193, 27, 216, 111, 98, 193, 222, 53, 196, 155, 50, 209, 227, 67, 238, 45, 223, 68, 25, 62, 63, 157, 164, 229, 140, 113, 93, 171, 186, 159, 1, 56, 240, 182, 182, 51, 48, 121, 203, 35, 89, 146, 128, 121, 203, 89, 179, 88, 59, 244, 251, 150, 41, 132, 134, 91, 86, 146, 227, 185, 179, 136, 76, 71, 200, 225, 41, 226, 208, 23, 97, 229, 107, 249, 156, 56, 17, 14, 55, 112, 55, 155, 230, 190, 91, 237, 44, 144, 248, 202, 246, 229, 64, 166, 230, 212, 60, 16, 65, 46, 79, 173, 157, 109, 207, 146, 133, 100, 123, 140, 7, 151, 104, 116, 82, 164, 165, 76, 235, 130, 190, 66, 105, 214, 74, 19, 127, 160, 188, 242, 34, 145, 28, 121, 241, 108, 126, 161, 35, 242, 155, 72, 47, 38, 236, 251, 89, 164, 190, 79, 228, 233, 161, 14, 31, 244, 61, 16, 231, 178, 33, 212, 213, 113, 36, 134, 89, 52, 123, 248, 204, 255, 211, 212, 164, 5, 144, 109, 8, 9, 60, 214, 83, 254, 170, 204, 66, 239, 174, 147, 142, 75, 31, 125, 125, 154, 32, 143, 161, 26, 178, 220, 183, 114, 220, 142, 31, 201, 92, 15, 152, 248, 108, 100, 24, 146, 180, 209, 8, 22, 192, 168, 252, 0, 57, 101, 1, 3, 204, 109, 255, 217, 175, 188, 160, 223, 86, 106, 84, 16, 51, 152, 8, 76, 102, 90, 146, 123, 48, 28, 219, 78, 218, 182, 150, 124, 115, 227, 164, 231, 152, 42, 66, 175, 34, 172, 1, 194, 173, 144, 227, 248, 160, 238, 175, 156, 62, 203, 238, 102, 251, 65, 133, 88, 164, 63, 24, 215, 131, 47, 254, 207, 25, 119, 244, 151, 148, 57, 102, 142, 236, 72, 67, 196, 165, 6, 124, 118, 145, 99, 215, 2, 9, 118, 236, 136, 40, 205, 48, 220, 228, 14, 232, 13, 163, 119, 79, 149, 97, 37, 105, 35, 33, 47, 54, 200, 117, 180, 53, 240, 19, 74, 184, 36, 187, 1, 23, 181, 54, 69, 85, 185, 98, 178, 180, 198, 68, 12, 151, 198, 198, 146, 122, 0, 109, 248, 246, 147, 141, 126, 71, 222, 118, 42, 242, 55, 193, 40, 133, 253, 178, 137, 188, 137, 179, 192, 216, 152, 165, 113, 255, 68, 23, 55, 54, 232, 17, 181, 174, 50, 224, 104, 199, 151, 1, 233, 198, 224, 239, 0, 74, 66, 167, 102, 80, 7, 54, 211, 156, 63, 169, 45, 42, 247, 43, 178, 252, 250, 8, 2, 106, 193, 90, 15, 100, 51, 94, 92, 198, 160, 83, 13, 16, 106, 0, 52, 180, 3, 159, 216, 122, 116, 83, 171, 13, 4, 34, 130, 107, 190, 201, 92, 254, 64, 108, 3, 130, 99, 190, 154, 57, 7, 113, 72, 221, 232, 247, 100, 178, 189, 29, 19, 2, 3, 1, 0, 1}

// const idProtocol = "https"
// const idHostname = "id.sage.party"
const idProtocol = "http"
const idHostname = "localhost:3001"

func (b *Bindings) InitId(email string) error {
	if identity == nil {
		return errors.New("Identity not loaded")
	}
	if currentUrl == nil || currentUrl.Host == idHostname {
		return fmt.Errorf("this binding must not be called from %s", idHostname)
	}
	intent := make([]byte, 16)
	if _, err := rand.Read(intent); err != nil {
		return err
	}

	clientIntent = base64.RawURLEncoding.EncodeToString(intent)
	idEmail = email

	query := url.Values{
		"intent": {clientIntent},
		"email":  {email},
	}
	b.ui.Navigate(fmt.Sprintf("%s://%s/register?%s", idProtocol, idHostname, query.Encode()))
	return nil
}

func (b *Bindings) FinalizeId(intent string, serverIntent string, password string) (string, error) {
	// prevent bruteforce and replay attacks
	defer func() {
		clientIntent = ""
		idEmail = ""
	}()

	if currentUrl == nil || currentUrl.Host != idHostname {
		return "", fmt.Errorf("this binding must be called from %s", idHostname)
	}
	if intent != clientIntent || clientIntent == "" {
		return "", errors.New("Invalid intent")
	}

	intentPK, err := x509.ParsePKIXPublicKey(serverECIntentPublicKey)
	if err != nil {
		return "", err
	}

	parsed, err := jwt.ParseSigned(serverIntent, []jose.SignatureAlgorithm{jose.ES256})
	if err != nil {
		return "", err
	}

	var claims jwt.Claims
	if err := parsed.Claims(intentPK, &claims); err != nil {
		return "", err
	}

	// we must make sure that the server knows the correct email
	// the ID isnt that important here since its validated at the start, but why not
	if err := claims.Validate(jwt.Expected{
		Issuer:      "v4",
		AnyAudience: jwt.Audience{"v4/id/server-intent"},
		ID:          clientIntent,
		Subject:     idEmail,
	}); err != nil {
		return "", err
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

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
	data := make([]byte, len(salt)+len(plaintext))
	copy(data, salt)
	gcm.Seal(data[len(salt):], salt[:gcm.NonceSize()], plaintext, nil)

	// to prevent a compromise of id.sage.party just immediately decrypting the key
	// using the password it knows, we encrypt the key with the server's public key as well
	// then the only way to decrypt the key is to have the password and the server's private key or
	// to decrypt the key after its been sent to the server
	// however that will require email verification (or server/database compromise ig?)

	rsaKey, err := x509.ParsePKCS1PublicKey(serverRSAIntentPublicKey)
	if err != nil {
		return "", err
	}
	sealedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, data, nil)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(sealedData), nil
}

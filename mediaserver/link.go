package mediaserver

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	errInvalidMsgLen = errors.New("invalid message length")
	errInvalidHMAC   = errors.New("invalid hmac")
)

var key = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

func (m *NetworkManager) handleV1Request(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if err := m.handleV1RequestInner(w, r); err != nil {
		m.logger.Error("V1 request error", "err", err)

		m.fallbackRoute(w, r)
	}
}

func (m *NetworkManager) handleV1RequestInner(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	rawMsg, err := decryptMessage(key, body)
	if err != nil {
		return err
	}

	var msg map[string]any

	if err := msgpack.Unmarshal(rawMsg, &msg); err != nil {
		return err
	}

	m.logger.Info("Msg", "msg", msg)

	resp := map[string]any{}

	resp["hello"] = "world"

	rawResp, err := msgpack.Marshal(&resp)
	if err != nil {
		return err
	}

	encResp, err := encryptMessage(key, rawResp)
	if err != nil {
		return err
	}

	_, err = w.Write(encResp)
	if err != nil {
		return err
	}

	return nil
}

func decryptMessage(key []byte, msg []byte) ([]byte, error) {
	if len(msg) < 16+32+1 {
		return nil, errInvalidMsgLen
	}

	var (
		iv         = msg[0:aes.BlockSize]
		cipherText = msg[aes.BlockSize:(len(msg) - 32)]
		mac        = msg[(len(msg) - 32):]
	)

	h := hmac.New(sha256.New, key)
	h.Write(cipherText)

	// Validate hmac
	if !hmac.Equal(mac, h.Sum(nil)) {
		return nil, errInvalidHMAC
	}

	// Decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(cipherText))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, cipherText)

	return plaintext, nil
}

func encryptMessage(key []byte, msg []byte) ([]byte, error) {
	// Generate iv
	iv := make([]byte, aes.BlockSize)

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(msg))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, msg)

	// Generate hmac
	h := hmac.New(sha256.New, key)
	h.Write(ciphertext)
	hmacSum := h.Sum(nil)

	combined := iv
	combined = append(combined, ciphertext...)
	combined = append(combined, hmacSum...)

	return combined, nil
}

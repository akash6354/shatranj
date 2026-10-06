package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	body := []byte(`{"event":"payment.captured"}`)
	mac := hmac.New(sha256.New, []byte("webhook-secret"))
	_, _ = mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))
	if !verifySignature("webhook-secret", body, signature) {
		t.Fatal("valid webhook signature rejected")
	}
	if verifySignature("wrong-secret", body, signature) {
		t.Fatal("signature accepted with wrong secret")
	}
	if verifySignature("webhook-secret", append(body, ' '), signature) {
		t.Fatal("signature accepted for modified body")
	}
}

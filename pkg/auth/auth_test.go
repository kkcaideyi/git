package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
)

func TestParseRejectsUnexpectedAlgorithm(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: "u", Role: "user"})
	raw, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(raw, "secret"); err == nil {
		t.Fatal("expected none algorithm to be rejected")
	}
}
func TestGenerateParseRoundTrip(t *testing.T) {
	raw, err := Generate("user-1", "user", "secret", 1)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Parse(raw, "secret")
	if err != nil || claims.UserID != "user-1" || claims.Role != "user" {
		t.Fatalf("round trip failed: %#v %v", claims, err)
	}
}

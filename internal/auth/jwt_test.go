package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() returned an error: %v", err)
	}

	if gotID != userID {
		t.Errorf("ValidateJWT() = %v, want %v", gotID, userID)
	}
}

func TestValidateJWTExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("ValidateJWT() expected an error for an expired token")
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("ValidateJWT() expected an error for the wrong secret")
	}
}

func TestValidateJWTInvalidToken(t *testing.T) {
	_, err := ValidateJWT("this-is-not-a-jwt", "test-secret")
	if err == nil {
		t.Fatal("ValidateJWT() expected an error for an invalid token")
	}
}

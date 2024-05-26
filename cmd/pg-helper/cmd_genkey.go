package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

type GenKeyCmd struct {
	PublicKey  string `short:"p" long:"publickey" default:"public.pem" help:"Public key file name"`
	PrivateKey string `short:"k" long:"privatekey" default:"private.pem" help:"Private key file name"`
	Force      bool   `short:"f" long:"force" default:"false" help:"Force to overwrite the existing key files"`
}

func (g *GenKeyCmd) Run() error {
	if !g.Force {
		if _, err := os.Stat(g.PublicKey); err == nil {
			err := errors.New("public key file already exists")
			log.Error().Err(err).Str("PublicKey", g.PublicKey).Msg("Failed to generate key pair")
			return err
		}
		if _, err := os.Stat(g.PrivateKey); err == nil {
			err := errors.New("private key file already exists")
			log.Error().Err(err).Str("PrivateKey", g.PrivateKey).Msg("Failed to generate key pair")
			return err
		}
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate key pair")
		return err
	}

	pkcs8PrivateKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal private key to PKCS8 format")
		return err
	}
	// Encode the private key to PEM format
	privateKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8PrivateKey,
	})

	pkixPublicKey, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	// Encode the public key to PEM format
	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixPublicKey,
	})

	// Write the private key to a file
	err = os.WriteFile(g.PrivateKey, privateKeyBytes, 0600)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write private key to file")
		return err
	}

	// Write the public key to a file
	err = os.WriteFile(g.PublicKey, publicKeyBytes, 0600)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write public key to file")
		return err
	}

	fmt.Println("Successfully generated Ed25519 key pair in PEM format")
	return nil
}

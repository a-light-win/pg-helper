package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/a-light-win/pg-helper/internal/validate"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type GenJwtCmd struct {
	Signer    string `validate:"required,file" help:"Path to the signer private key file"`
	Subject   string `help:"Name of the user or login entity"`
	Audiences []string
	Scopes    []string
	Resources []string
	Duration  time.Duration `help:"The duration of the token" default:"24h"`
	Output    string        `help:"The output file" type:"path"`
}

type customClaims struct {
	jwt.RegisteredClaims
	Scope    string `json:"scope,omitempty"`
	Resource string `json:"resource,omitempty"`
}

func (g *GenJwtCmd) Run(ctx *Context) error {
	validator := validate.New()
	if err := validator.Struct(g); err != nil {
		log.Error().Err(err).Msg("config validation failed")
		return err
	}

	if g.Subject == "" {
		g.Subject = uuid.New().String()
	}

	// Define the token claims
	claims := &customClaims{
		Scope:    strings.Join(g.Scopes, " "),
		Resource: strings.Join(g.Resources, " "),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   g.Subject,
			Audience:  g.Audiences,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(g.Duration)),
		},
	}

	key, err := os.ReadFile(g.Signer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read the signer private key file")
		return err
	}

	signer, err := jwt.ParseEdPrivateKeyFromPEM([]byte(key))
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse the signer private key")
		return err
	}

	// Create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	// Sign the token with a secret
	signedToken, err := token.SignedString(signer)
	if err != nil {
		return err
	}

	if g.Output != "" {
		err = os.WriteFile(g.Output, []byte(signedToken), 0600)
		if err != nil {
			log.Error().Err(err).Msg("Failed to write the signed token to the file")
			return err
		}
	} else {
		// Print the signed token
		fmt.Println(signedToken)
	}

	return nil
}

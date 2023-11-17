package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/urfave/cli/v2"
)

type Bot struct {
	Pin        string `json:"pin"`
	ClientID   string `json:"client_id"`
	SessionID  string `json:"session_id"`
	PinToken   string `json:"pin_token"`
	PrivateKey string `json:"private_key"`
}

func main() {
	app := &cli.App{
		Name:    "mixin-bot",
		Usage:   "Mixin bot API cli",
		Version: "2.0.1",
		Commands: []*cli.Command{
			{
				Name:   "transfer",
				Action: transferCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "asset,a",
						Usage: "asset",
					},
					&cli.StringFlag{
						Name:  "amount,z",
						Usage: "amount",
					},
					&cli.StringFlag{
						Name:  "receiver,r",
						Usage: "receiver",
					},
					&cli.StringFlag{
						Name:  "keystore,k",
						Usage: "keystore download from https://developers.mixin.one/dashboard",
					},
				},
			},
			{
				Name:   "migrate",
				Action: botMigrateTIPCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "keystore,k",
						Usage: "keystore download from https://developers.mixin.one/dashboard",
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func transferCmd(c *cli.Context) error {
	keystore := c.String("keystore")
	asset := c.String("asset")
	amount := c.String("amount")
	receiver := c.String("receiver")

	dat, err := os.ReadFile(keystore)
	if err != nil {
		panic(err)
	}
	var user Bot
	err = json.Unmarshal([]byte(dat), &user)
	if err != nil {
		panic(err)
	}

	su := &bot.SafeUser{
		UserId:     user.ClientID,
		SessionId:  user.SessionID,
		SessionKey: user.PrivateKey,
		SpendKey:   user.Pin[:64],
	}

	ma := bot.NewUUIDMixAddress([]string{receiver}, 1)
	tr := &bot.TransactionRecipient{MixAddress: ma.String(), Amount: amount}
	trace := bot.UuidNewV4().String()
	log.Println("trace:", trace)
	tx, err := bot.SendTransaction(context.Background(), asset, []*bot.TransactionRecipient{tr}, trace, su)
	if err != nil {
		return err
	}
	log.Println("tx:", tx.PayloadHash().String())
	return nil
}

func botMigrateTIPCmd(c *cli.Context) error {
	keystore := c.String("keystore")

	dat, err := os.ReadFile(keystore)
	if err != nil {
		panic(err)
	}
	var app Bot
	err = json.Unmarshal([]byte(dat), &app)
	if err != nil {
		panic(err)
	}

	tipPub, tipPriv, _ := ed25519.GenerateKey(rand.Reader)
	log.Printf("Your tip private key: %s", hex.EncodeToString(tipPriv))

	err = bot.UpdateTipPin(context.Background(), app.Pin, hex.EncodeToString(tipPub), app.PinToken, app.ClientID, app.SessionID, app.PrivateKey)
	if err != nil {
		return fmt.Errorf("bot.UpdateTipPin() => %v", err)
	}

	app.Pin = hex.EncodeToString(tipPriv)
	keystoreRaw, _ := json.Marshal(app)
	log.Printf("your new keystore after migrate: %s", string(keystoreRaw))
	return nil
}

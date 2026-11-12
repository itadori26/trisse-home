package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

// Output of the setup-status command, used by the native package's helper
// scripts (packaging/) to decide whether the setup token is still needed.
const (
	setupPending = "pending"
	setupDone    = "done"
)

// setupStatus prints "pending" while no account exists and "done" afterwards.
// It asks the running server on loopback (always allowed by the IP allowlist),
// so it needs neither the database nor the secret key.
func setupStatus() error {
	base, err := loopbackURL(os.Getenv("LISTEN_ADDR"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()
	status, err := fetchSetupStatus(ctx, http.DefaultClient, base)
	if err != nil {
		return err
	}
	fmt.Println(status)
	return nil
}

// loopbackURL returns the base URL of the local server for LISTEN_ADDR.
func loopbackURL(listenAddr string) (string, error) {
	if listenAddr == "" {
		listenAddr = ":8090"
	}
	_, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return "", fmt.Errorf("parse LISTEN_ADDR: %w", err)
	}
	return "http://127.0.0.1:" + port, nil
}

func fetchSetupStatus(ctx context.Context, client *http.Client, baseURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/auth/status", nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("server not reachable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", errors.New("unexpected status: " + res.Status)
	}
	var body struct {
		SetupRequired *bool `json:"setupRequired"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&body); err != nil {
		return "", fmt.Errorf("decode status: %w", err)
	}
	if body.SetupRequired == nil {
		return "", errors.New("decode status: setupRequired missing")
	}
	if *body.SetupRequired {
		return setupPending, nil
	}
	return setupDone, nil
}

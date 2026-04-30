package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

var KeyLength = 32

func main() {
	key := make([]byte, KeyLength)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	secret := base64.StdEncoding.EncodeToString(key)

	envFile := ".env"
	content, _ := os.ReadFile(envFile)
	lines := bytes.Split(content, []byte("\n"))

	var found bool
	for i, line := range lines {
		if bytes.HasPrefix(line, []byte("JWT_SECRET=")) {
			lines[i] = []byte("JWT_SECRET=" + secret)
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, []byte("JWT_SECRET="+secret))
	}

	f, err := os.Create(envFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		if len(strings.TrimSpace(string(line))) > 0 {
			fmt.Fprintln(w, string(line))
		}
	}
	w.Flush()

	fmt.Println("✅ New JWT_SECRET saved to .env")
}

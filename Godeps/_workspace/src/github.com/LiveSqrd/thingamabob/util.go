package thingamabob

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"gopkg.in/redis.v2"
)

type script struct {
	text string
	sha1 string
}

func newScript(text string) script {
	sha1Raw := sha1.Sum([]byte(text))
	sha1 := hex.EncodeToString(sha1Raw[:])
	return script{text, sha1}
}

func (s script) eval(client *redis.Client, keys []string, args []string) *redis.Cmd {
	cmd := client.EvalSha(s.sha1, keys, args)
	if err := cmd.Err(); err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		return client.Eval(s.text, keys, args)
	}
	return cmd
}

package ratelimiter

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/go-redis/redis/v8"
)

// LoadLuaScript reads the lua file at path and loads it into Redis, returning the SHA.
func LoadLuaScript(ctx context.Context, rdb *redis.Client, relativePath string) (string, error) {
	path := filepath.Clean(relativePath)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	sha, err := rdb.ScriptLoad(ctx, string(b)).Result()
	if err != nil {
		return "", err
	}
	log.Printf("Loaded Lua script %s -> SHA %s\n", path, sha)
	return sha, nil
}

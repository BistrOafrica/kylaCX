package service

import "sync"

// TokenBlacklist is a struct to manage blacklisted tokens.
type TokenBlacklist struct {
	blacklistedTokens map[string]struct{}
	mu                sync.Mutex
}

// NewTokenBlacklist creates a new TokenBlacklist.
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		blacklistedTokens: make(map[string]struct{}),
	}
}

// AddToken adds a token to the blacklist.
func (tb *TokenBlacklist) AddToken(token string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.blacklistedTokens[token] = struct{}{}
}

// IsTokenBlacklisted checks if a token is blacklisted.
func (tb *TokenBlacklist) IsTokenBlacklisted(token string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	_, blacklisted := tb.blacklistedTokens[token]
	return blacklisted
}

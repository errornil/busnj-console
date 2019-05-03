package applemaps

// TokenGenerator represents JWT token generator for Apple Maps
type TokenGenerator struct {
	pathToKey string
	keyID     string
}

// NewTokenGenerator creates new TokenGenerator
func NewTokenGenerator(pathToKey, keyID string) (*TokenGenerator, error) {
	return &TokenGenerator{
		pathToKey: pathToKey,
		keyID:     keyID,
	}, nil
}

// GetToken generates new JWT token
func (tg *TokenGenerator) GetToken() string {
	return "TOKEN"
}

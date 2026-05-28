package auth

import "kyla-be/pkg/service"

// JWTManager is a re-export of the canonical JWTManager from the service layer.
type JWTManager = service.JWTManager

// DBAuthStore is a re-export of the session/device DB store from the service layer.
type DBAuthStore = service.DBAuthStore

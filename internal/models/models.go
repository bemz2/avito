package models

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Coins        int       `json:"coins" db:"coins"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Item struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Price     int       `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID              int       `json:"id" db:"id"`
	FromUserID      *int      `json:"from_user_id" db:"from_user_id"`
	ToUserID        *int      `json:"to_user_id" db:"to_user_id"`
	Amount          int       `json:"amount" db:"amount"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type InfoResponse struct {
	Coins       int                    `json:"coins"`
	Inventory   []InventoryItem        `json:"inventory"`
	CoinHistory CoinTransactionHistory `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinTransactionHistory struct {
	Received []CoinTransaction `json:"received"`
	Sent     []CoinTransaction `json:"sent"`
}

type CoinTransaction struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

package repository

import (
	"avito-shop/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("не найдено")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (username, password_hash, coins)
        VALUES ($1, $2, $3)
        RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		user.Username, user.PasswordHash, user.Coins,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, username, password_hash, coins, created_at
        FROM users WHERE username = $1`

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash,
		&user.Coins, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к БД: %w", err)
	}

	return user, nil
}

func (r *Repository) TransferCoins(ctx context.Context, fromUserID, toUserID int, amount int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var senderCoins int
	err = tx.QueryRowContext(ctx,
		"SELECT coins FROM users WHERE id = $1 FOR UPDATE",
		fromUserID,
	).Scan(&senderCoins)
	if err != nil {
		return err
	}

	if senderCoins < amount {
		return errors.New("недостаточно монет")
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET coins = coins - $1 WHERE id = $2",
		amount, fromUserID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET coins = coins + $1 WHERE id = $2",
		amount, toUserID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type)
         VALUES ($1, $2, $3, 'transfer')`,
		fromUserID, toUserID, amount,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) BuyItem(ctx context.Context, userID int, itemID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var itemPrice, userCoins int
	err = tx.QueryRowContext(ctx,
		"SELECT price FROM items WHERE id = $1",
		itemID,
	).Scan(&itemPrice)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		"SELECT coins FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&userCoins)
	if err != nil {
		return err
	}

	if userCoins < itemPrice {
		return errors.New("недостаточно монет")
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET coins = coins - $1 WHERE id = $2",
		itemPrice, userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_items (user_id, item_id, quantity)
         VALUES ($1, $2, 1)
         ON CONFLICT (user_id, item_id)
         DO UPDATE SET quantity = user_items.quantity + 1`,
		userID, itemID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (from_user_id, amount, transaction_type)
         VALUES ($1, $2, 'purchase')`,
		userID, itemPrice,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetUserInfo(ctx context.Context, userID int) (*models.InfoResponse, error) {
	info := &models.InfoResponse{}

	err := r.db.QueryRowContext(ctx,
		"SELECT coins FROM users WHERE id = $1",
		userID,
	).Scan(&info.Coins)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT i.name, ui.quantity
         FROM user_items ui
         JOIN items i ON ui.item_id = i.id
         WHERE ui.user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info.Inventory = []models.InventoryItem{}
	for rows.Next() {
		var item models.InventoryItem
		err := rows.Scan(&item.Type, &item.Quantity)
		if err != nil {
			return nil, err
		}
		info.Inventory = append(info.Inventory, item)
	}

	receivedRows, err := r.db.QueryContext(ctx,
		`SELECT u.username, t.amount
         FROM transactions t
         JOIN users u ON t.from_user_id = u.id
         WHERE t.to_user_id = $1 AND t.transaction_type = 'transfer'`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer receivedRows.Close()

	info.CoinHistory.Received = []models.CoinTransaction{}
	for receivedRows.Next() {
		var transaction models.CoinTransaction
		err := receivedRows.Scan(&transaction.FromUser, &transaction.Amount)
		if err != nil {
			return nil, err
		}
		info.CoinHistory.Received = append(info.CoinHistory.Received, transaction)
	}

	sentRows, err := r.db.QueryContext(ctx,
		`SELECT u.username, t.amount
         FROM transactions t
         JOIN users u ON t.to_user_id = u.id
         WHERE t.from_user_id = $1 AND t.transaction_type = 'transfer'`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer sentRows.Close()

	info.CoinHistory.Sent = []models.CoinTransaction{}
	for sentRows.Next() {
		var transaction models.CoinTransaction
		err := sentRows.Scan(&transaction.ToUser, &transaction.Amount)
		if err != nil {
			return nil, err
		}
		info.CoinHistory.Sent = append(info.CoinHistory.Sent, transaction)
	}

	return info, nil
}

func (r *Repository) GetItemByName(ctx context.Context, name string) (*models.Item, error) {
	item := &models.Item{}
	query := `
        SELECT id, name, price, created_at
        FROM items WHERE name = $1`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&item.ID, &item.Name, &item.Price, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("товар не найден")
	}
	return item, err
}

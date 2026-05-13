package sqlite

import (
	"context"
	"database/sql"
	"errors"
)

type ItemQRTokenReservationRepo struct{ db DBConn }

func NewItemQRTokenReservationRepo(db *sql.DB) *ItemQRTokenReservationRepo {
	return &ItemQRTokenReservationRepo{db: db}
}

func NewItemQRTokenReservationRepoConn(c DBConn) *ItemQRTokenReservationRepo {
	return &ItemQRTokenReservationRepo{db: c}
}

func (r *ItemQRTokenReservationRepo) Reserve(ctx context.Context, itemID, token string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO item_qr_token_reservations (item_id, qr_token)
		VALUES (?, ?)`,
		itemID, token)
	return err
}

func (r *ItemQRTokenReservationRepo) GetTokenByItemID(ctx context.Context, itemID string) (string, bool, error) {
	var token string
	err := r.db.QueryRowContext(ctx, `
		SELECT qr_token FROM item_qr_token_reservations WHERE item_id = ?`, itemID).Scan(&token)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return token, true, nil
}

func (r *ItemQRTokenReservationRepo) GetItemIDByToken(ctx context.Context, token string) (string, bool, error) {
	var itemID string
	err := r.db.QueryRowContext(ctx, `
		SELECT item_id FROM item_qr_token_reservations WHERE qr_token = ?`, token).Scan(&itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return itemID, true, nil
}

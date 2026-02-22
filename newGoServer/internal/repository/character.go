package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Character mirrors the characters table.
type Character struct {
	ID       int32
	OwnerID  int32
	Slot     int16
	Nickname string
	Class    int16
	Gender   int16
	Head     int16
	Face     int16
	Body     int16
	HomeMapNo int32
	HomePosX float32
	HomePosY float32
	HomePosZ float32
	RegDate  time.Time
	DelDate  *time.Time
}

// PcState mirrors the pc_state table.
type PcState struct {
	CharacterID  int32
	Level        int16
	Exp          int64
	Hp           int32
	Mp           int32
	Strength     int32
	Intelligence int32
	Dexterity    int32
	MapNo        int32
	PosX         float32
	PosY         float32
	PosZ         float32
	Stomach      int16
	PkCount      int32
	Chaotic      int32
}

// EquippedItem maps an equipment slot to an item.
type EquippedItem struct {
	Slot     int16
	SerialNo int64
	ItemNo   int32
}

// CharacterRepository provides data access for characters and related tables.
type CharacterRepository struct {
	db *pgxpool.Pool
}

func NewCharacterRepository(db *pgxpool.Pool) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// GetByOwner returns all (non-deleted) characters for an account, ordered by slot.
func (r *CharacterRepository) GetByOwner(ctx context.Context, ownerID int32) ([]*Character, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, owner_id, slot, nickname, class, gender, head, face, body,
		        home_map_no, home_pos_x, home_pos_y, home_pos_z, reg_date
		 FROM characters
		 WHERE owner_id = $1 AND del_date IS NULL
		 ORDER BY slot`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []*Character
	for rows.Next() {
		c := &Character{}
		if err := rows.Scan(
			&c.ID, &c.OwnerID, &c.Slot, &c.Nickname, &c.Class, &c.Gender,
			&c.Head, &c.Face, &c.Body,
			&c.HomeMapNo, &c.HomePosX, &c.HomePosY, &c.HomePosZ, &c.RegDate,
		); err != nil {
			return nil, err
		}
		chars = append(chars, c)
	}
	return chars, rows.Err()
}

// GetByID fetches a single character. Returns nil if not found or deleted.
func (r *CharacterRepository) GetByID(ctx context.Context, id int32) (*Character, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, owner_id, slot, nickname, class, gender, head, face, body,
		        home_map_no, home_pos_x, home_pos_y, home_pos_z, reg_date
		 FROM characters WHERE id = $1 AND del_date IS NULL`,
		id,
	)
	c := &Character{}
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.Slot, &c.Nickname, &c.Class, &c.Gender,
		&c.Head, &c.Face, &c.Body,
		&c.HomeMapNo, &c.HomePosX, &c.HomePosY, &c.HomePosZ, &c.RegDate,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

// GetState returns the mutable state for a character.
func (r *CharacterRepository) GetState(ctx context.Context, charID int32) (*PcState, error) {
	row := r.db.QueryRow(ctx,
		`SELECT character_id, level, exp, hp, mp, strength, intelligence, dexterity,
		        map_no, pos_x, pos_y, pos_z, stomach, pk_count, chaotic
		 FROM pc_state WHERE character_id = $1`,
		charID,
	)
	s := &PcState{}
	err := row.Scan(
		&s.CharacterID, &s.Level, &s.Exp, &s.Hp, &s.Mp,
		&s.Strength, &s.Intelligence, &s.Dexterity,
		&s.MapNo, &s.PosX, &s.PosY, &s.PosZ,
		&s.Stomach, &s.PkCount, &s.Chaotic,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

// GetEquipped returns all equipped items for a character.
func (r *CharacterRepository) GetEquipped(ctx context.Context, charID int32) ([]EquippedItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT e.slot, e.serial_no, i.item_no
		 FROM pc_equip e JOIN pc_inventory i ON i.serial_no = e.serial_no
		 WHERE e.character_id = $1`,
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EquippedItem
	for rows.Next() {
		var item EquippedItem
		if err := rows.Scan(&item.Slot, &item.SerialNo, &item.ItemNo); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Create inserts a new character and its initial pc_state. Returns the new character id.
func (r *CharacterRepository) Create(ctx context.Context, c *Character, initialHP, initialMP int32) (int32, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var charID int32
	err = tx.QueryRow(ctx,
		`INSERT INTO characters (owner_id, slot, nickname, class, gender, head, face, body,
		                         home_map_no, home_pos_x, home_pos_y, home_pos_z)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		 RETURNING id`,
		c.OwnerID, c.Slot, c.Nickname, c.Class, c.Gender, c.Head, c.Face, c.Body,
		c.HomeMapNo, c.HomePosX, c.HomePosY, c.HomePosZ,
	).Scan(&charID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO pc_state (character_id, level, exp, hp, mp,
		                       strength, intelligence, dexterity,
		                       map_no, pos_x, pos_y, pos_z)
		 VALUES ($1, 1, 0, $2, $3, 10, 10, 10, $4, $5, $6, $7)`,
		charID, initialHP, initialMP,
		c.HomeMapNo, c.HomePosX, c.HomePosY, c.HomePosZ,
	)
	if err != nil {
		return 0, err
	}

	return charID, tx.Commit(ctx)
}

// SoftDelete marks a character as deleted without removing the row.
func (r *CharacterRepository) SoftDelete(ctx context.Context, charID, ownerID int32) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE characters SET del_date = NOW()
		 WHERE id = $1 AND owner_id = $2 AND del_date IS NULL`,
		charID, ownerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("character not found or already deleted")
	}
	return nil
}

// RecordLogin updates last_login timestamp in pc_state.
func (r *CharacterRepository) RecordLogin(ctx context.Context, charID int32) error {
	_, err := r.db.Exec(ctx,
		`UPDATE pc_state SET last_login = NOW() WHERE character_id = $1`,
		charID,
	)
	return err
}

// SaveState persists position and stats on logout.
func (r *CharacterRepository) SaveState(ctx context.Context, s *PcState) error {
	_, err := r.db.Exec(ctx,
		`UPDATE pc_state
		 SET level=$2, exp=$3, hp=$4, mp=$5,
		     strength=$6, intelligence=$7, dexterity=$8,
		     map_no=$9, pos_x=$10, pos_y=$11, pos_z=$12,
		     stomach=$13, pk_count=$14, chaotic=$15,
		     last_logout=NOW()
		 WHERE character_id=$1`,
		s.CharacterID, s.Level, s.Exp, s.Hp, s.Mp,
		s.Strength, s.Intelligence, s.Dexterity,
		s.MapNo, s.PosX, s.PosY, s.PosZ,
		s.Stomach, s.PkCount, s.Chaotic,
	)
	return err
}

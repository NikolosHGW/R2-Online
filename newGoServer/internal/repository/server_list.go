package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"r2server/internal/packet/login/send"
)

// GameServerRepository provides data access for game_servers table.
type GameServerRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewGameServerRepository(db *pgxpool.Pool, rdb *redis.Client) *GameServerRepository {
	return &GameServerRepository{db: db, redis: rdb}
}

// GetAll returns all game servers sorted by server_id.
func (r *GameServerRepository) GetAll(ctx context.Context) ([]send.GameServer, error) {
	rows, err := r.db.Query(ctx,
		`SELECT server_id, name, server_ip, server_port, type, hidden, status, congestion
		 FROM game_servers ORDER BY server_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []send.GameServer
	for rows.Next() {
		var s send.GameServer
		var hidden, status bool
		if err := rows.Scan(
			&s.ServerID, &s.Name, &s.IP, &s.Port,
			&s.Type, &hidden, &status, &s.Congestion,
		); err != nil {
			return nil, err
		}
		s.Hidden = hidden
		s.Status = status
		servers = append(servers, s)
	}
	return servers, rows.Err()
}

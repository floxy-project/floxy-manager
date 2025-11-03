package membershipaudit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rom8726/floxy-manager/pkg/db"
)

// Write inserts a record into membership_audit within the current transaction (if any).
// oldVal and newVal are marshaled to JSON. If nil, corresponding column gets NULL.
func Write(
	ctx context.Context,
	exec db.Tx,
	membershipID string,
	actorUserID int,
	action string,
	oldVal any,
	newVal any,
) error {
	var (
		oldJSON []byte
		newJSON []byte
		err     error
	)

	if oldVal != nil {
		oldJSON, err = json.Marshal(oldVal)
		if err != nil {
			return fmt.Errorf("marshal old value: %w", err)
		}
	}

	if newVal != nil {
		newJSON, err = json.Marshal(newVal)
		if err != nil {
			return fmt.Errorf("marshal new value: %w", err)
		}
	}

	const query = `
		insert into workflows_manager.membership_audit (membership_id, actor_user_id, action, old_value, new_value)
		values ($1, $2, $3, $4, $5)
	`

	if _, err := exec.Exec(ctx, query, membershipID, actorUserID, action, oldJSON, newJSON); err != nil {
		return fmt.Errorf("insert membership_audit: %w", err)
	}

	return nil
}

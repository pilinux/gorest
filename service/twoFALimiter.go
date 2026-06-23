package service

import (
	"context"
	"strconv"
	"time"

	"github.com/mediocregopher/radix/v4"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
)

// Backup-code brute-force limiter tuning.
//
// The first backup2FAFreeAttempts failures are allowed with no cooldown. The
// failure that exhausts them triggers a fixed backup2FABaseCooldownMin cooldown;
// every consecutive failure after that escalates to backup2FAStepCooldownMin ×
// (total attempts) minutes. The limiter state is kept for backup2FAResetWindow
// beyond the current cooldown, so an account that stops failing eventually
// resets to zero.
const (
	backup2FAFreeAttempts    = 3
	backup2FABaseCooldownMin = 10
	backup2FAStepCooldownMin = 5
	backup2FAResetWindow     = time.Hour
)

// backup2FALimitScript increments the failed-attempt counter, derives the
// resulting cooldown from the escalating schedule, stores the cooldown deadline,
// and refreshes the key TTL — all atomically. It returns the cooldown length in
// seconds (0 while still within the free attempts).
//
// KEYS[1]=limiter key.
// ARGV[1]=attempts field, ARGV[2]=cooldown-until field,
// ARGV[3]=free attempts, ARGV[4]=base cooldown min, ARGV[5]=step cooldown min,
// ARGV[6]=reset window seconds, ARGV[7]=current unix time.
var backup2FALimitScript = radix.NewEvalScript(`
local attempts = redis.call('HINCRBY', KEYS[1], ARGV[1], 1)
local free = tonumber(ARGV[3])
local cooldown = 0
if attempts == free then
	cooldown = tonumber(ARGV[4]) * 60
elseif attempts > free then
	cooldown = tonumber(ARGV[5]) * attempts * 60
end
redis.call('HSET', KEYS[1], ARGV[2], tonumber(ARGV[7]) + cooldown)
redis.call('EXPIRE', KEYS[1], cooldown + tonumber(ARGV[6]))
return cooldown
`)

// backup2FALimitKey returns the Redis key for an account's backup-code limiter.
func backup2FALimitKey(authID uint64) string {
	return model.Backup2FALimitKeyPrefix + strconv.FormatUint(authID, 10)
}

// backup2FALimitCtx builds a context with the configured Redis connection TTL.
func backup2FALimitCtx() (context.Context, context.CancelFunc) {
	rConnTTL := config.GetConfig().Database.REDIS.Conn.ConnTTL
	return context.WithTimeout(context.Background(), time.Duration(rConnTTL)*time.Second)
}

// Backup2FACooldown reports how long an account must wait before its next
// backup-code attempt. It returns 0 when no cooldown is in effect (including
// when Redis is disabled, in which case the limiter is skipped).
func Backup2FACooldown(authID uint64) (remaining time.Duration, err error) {
	if !config.IsRedis() {
		return 0, nil
	}

	client := database.GetRedis()
	if client == nil {
		return 0, database.ErrRedisNotInitialized
	}
	ctx, cancel := backup2FALimitCtx()
	defer cancel()

	var until int64
	if err = client.Do(ctx, radix.FlatCmd(&until, "HGET",
		backup2FALimitKey(authID), model.Backup2FAFieldCooldownUntil)); err != nil {
		return 0, err
	}

	remaining = time.Until(time.Unix(until, 0))
	if remaining <= 0 {
		return 0, nil
	}
	return remaining, nil
}

// RegisterBackup2FAFailure records a failed backup-code attempt and returns the
// cooldown now imposed (0 while still within the free attempts). It is a no-op
// returning 0 when Redis is disabled.
func RegisterBackup2FAFailure(authID uint64) (cooldown time.Duration, err error) {
	if !config.IsRedis() {
		return 0, nil
	}

	client := database.GetRedis()
	if client == nil {
		return 0, database.ErrRedisNotInitialized
	}
	ctx, cancel := backup2FALimitCtx()
	defer cancel()

	var cooldownSec int64
	if err = client.Do(ctx, backup2FALimitScript.Cmd(&cooldownSec,
		[]string{backup2FALimitKey(authID)},
		model.Backup2FAFieldAttempts,
		model.Backup2FAFieldCooldownUntil,
		strconv.Itoa(backup2FAFreeAttempts),
		strconv.Itoa(backup2FABaseCooldownMin),
		strconv.Itoa(backup2FAStepCooldownMin),
		strconv.FormatInt(int64(backup2FAResetWindow/time.Second), 10),
		strconv.FormatInt(time.Now().Unix(), 10),
	)); err != nil {
		return 0, err
	}

	return time.Duration(cooldownSec) * time.Second, nil
}

// ClearBackup2FALimit drops an account's backup-code limiter state after a
// successful validation. It is a no-op when Redis is disabled.
func ClearBackup2FALimit(authID uint64) error {
	if !config.IsRedis() {
		return nil
	}

	client := database.GetRedis()
	if client == nil {
		return database.ErrRedisNotInitialized
	}
	ctx, cancel := backup2FALimitCtx()
	defer cancel()

	if err := client.Do(ctx, radix.FlatCmd(nil, "DEL", backup2FALimitKey(authID))); err != nil {
		return err
	}
	return nil
}

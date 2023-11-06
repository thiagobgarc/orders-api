package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/thiagobgarc/orders-api/model"
)

type RedisRepo struct {
	Client *redis.Client
}

// orderIDKey generates a key for an order ID.
//
// It takes an unsigned 64-bit integer `id` as a parameter.
// It returns a string representing the generated key.
func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

// Insert inserts an order into the RedisRepo.
//
// The function takes a context.Context and a model.Order as input parameters.
// It serializes the order into JSON format and inserts it into Redis using a key generated from the order ID.
// The function returns an error if there is any issue with encoding the order, inserting the order into Redis, or committing the transaction.
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order! %m", err)
	}

	key := orderIDKey(order.OrderID)

	txn := r.Client.TxPipeline()

	res := r.Client.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order! %m", err)
	}

	if err != txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order! %m", err)
	}

	if err := txn.Commit(ctx).Err(); err != nil {
		return fmt.Errorf("failed to insert order! %m", err)
	}

	return nil
}

var ErrNotExist = errors.New("order does not exist")

// FindByID finds an order by its ID.
//
// Parameters:
// - ctx: the context.Context for the operation.
// - id: the ID of the order to find.
//
// Returns:
// - model.Order: the found order.
// - error: an error if the order does not exist or if there was an error finding or decoding the order.
func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("failed to find order! %m", err)
	}

	var order model.Order
	err = json.Unmarshal(([]byte(value), &order))
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order! %m", err)
	}

	return order, nil
}

// DeletedByID deletes an order from Redis based on its ID.
//
// ctx: the context.Context object for cancellation and timeouts.
// id: the ID of the order to be deleted.
// error: returns an error if there was a problem deleting the order.
func (r *RedisRepo) DeletedByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	txn := r.Client.TxPipeline()

	err := r.Client.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to delete order! %m", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete order! %m", err)
	}

	if _. err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete order! %m", err)
	}

	return nil
}

// Update updates the RedisRepo with the given order.
//
// It takes a context.Context object and a model.Order object as parameters.
// It returns an error.
func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order! %m", err)
	}

	key := orderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update order! %m", err)
	}

	return nil
}

type FindALLPage struct {
	Size uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

// FindAll retrieves all the records from the RedisRepo.
//
// It takes the following parameters:
// - ctx: the context.Context object for handling cancellation and timeouts.
// - page: the FindAllPage object containing the page offset and size.
//
// It returns a FindResult object and an error if any occurred.
func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "x", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to find orders! %m", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to find orders! %m", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal(([]byte(x), &order))
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order! %m", err)
		}

		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}

// Package mgobarrier is designed to solve the timing problem of accessing
// RM(Resource Manager) based on MongoDB in distributed transactions.
package mgobarrier

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

// DefaultDBCollection is the default database name and collection name separated by ".".
var DefaultDBCollection = "tdxa.txbarrier"

type clientKey struct{}

// DoWithClient is a shortcut of DoWithSessionContext.
func DoWithClient(ctx context.Context, cli *mongo.Client, fn func(ctx context.Context) error) error {
	return cli.UseSession(ctx, func(ctx context.Context) error {
		newCtx := context.WithValue(ctx, clientKey{}, cli)
		return DoWithSessionContext(newCtx, fn)
	})
}

// DoWithSessionContext is used to solve the timing problem in distributed transactions.
// It returns txbarrier.ErrDuplicationOrSuspension or txbarrier.ErrEmptyCompensation if
// occurs duplicated request, empty compensation or hanging request.
func DoWithSessionContext(ctx context.Context, fn func(ctx context.Context) error) error {
	cli, ok := ctx.Value(clientKey{}).(*mongo.Client)
	if !ok {
		return errors.New("client is not provided")
	}

	sc, err := cli.StartSession()
	if err != nil {
		return err
	}
	err = sc.StartTransaction()
	if err != nil {
		return err
	}

	if b := txbarrier.BarrierFromCtx(ctx); b.Valid() { // check whether if need barrier check.
		err = barrierCheck(ctx, cli, b)
		if errors.Is(err, txbarrier.ErrEmptyCompensation) {
			_ = sc.CommitTransaction(ctx)
			return err
		}
		if err != nil {
			_ = sc.AbortTransaction(ctx)
			return err
		}
	}

	if err = fn(ctx); err == nil {
		err = sc.CommitTransaction(ctx)
	} else {
		err = multierror.Append(err, sc.AbortTransaction(ctx))
	}

	return err
}

func barrierCheck(ctx context.Context, cli *mongo.Client, b *txbarrier.Barrier) error {
	affected, err := insertDB(ctx, cli, b, b.Op, string(b.Op))
	if err != nil {
		return err
	}
	if affected == 0 { // duplicated or hanging request
		return txbarrier.ErrDuplicationOrSuspension
	}

	if b.Op == txbarrier.Cancel {
		affected, err = insertDB(ctx, cli, b, txbarrier.Try, string(b.Op))
		if err != nil {
			return err
		}
		if affected > 0 { // empty compensation
			return txbarrier.ErrEmptyCompensation
		}
	}

	return nil
}

func insertDB(ctx context.Context, cli *mongo.Client, b *txbarrier.Barrier, op txbarrier.Operation, reason string) (int64, error) {
	tmp := strings.Split(DefaultDBCollection, ".")
	if len(tmp) != 2 {
		return 0, fmt.Errorf("invalid db collection name `%s`", DefaultDBCollection)
	}

	_, err := cli.Database(tmp[0]).Collection(tmp[1]).InsertOne(ctx, bson.D{
		{Key: "xid", Value: b.XID},
		{Key: "branch_id", Value: b.BranchID},
		{Key: "op", Value: op},
		{Key: "reason", Value: reason},
	})
	if mongo.IsDuplicateKeyError(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return 1, nil
}

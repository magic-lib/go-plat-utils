// Package mgobarrier is designed to solve the timing problem of accessing
// RM(Resource Manager) based on MongoDB in distributed transactions.
package mgobarrier

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

// DefaultDBCollection is the default database name and collection name separated by ".".
var DefaultDBCollection = "tdxa.txbarrier"

// DoWithClient is a shortcut of DoWithSessionContext.
func DoWithClient(ctx context.Context, cli *mongo.Client, fn func(sc mongo.SessionContext) error) error {
	return cli.UseSession(ctx, func(sc mongo.SessionContext) error {
		return DoWithSessionContext(sc, fn)
	})
}

// DoWithSessionContext is used to solve the timing problem in distributed transactions.
// It returns txbarrier.ErrDuplicationOrSuspension or txbarrier.ErrEmptyCompensation if
// occurs duplicated request, empty compensation or hanging request.
func DoWithSessionContext(sc mongo.SessionContext, fn func(sc mongo.SessionContext) error) error {
	err := sc.StartTransaction()
	if err != nil {
		return err
	}

	if b := txbarrier.BarrierFromCtx(sc); b.Valid() { // check whether if need barrier check.
		err = barrierCheck(sc, b)
		if err == txbarrier.ErrEmptyCompensation {
			_ = sc.CommitTransaction(sc)
			return err
		}
		if err != nil {
			_ = sc.AbortTransaction(sc)
			return err
		}
	}

	if err = fn(sc); err == nil {
		err = sc.CommitTransaction(sc)
	} else {
		err = multierror.Append(err, sc.AbortTransaction(sc))
	}

	return err
}

func barrierCheck(sc mongo.SessionContext, b *txbarrier.Barrier) error {
	affected, err := insertDB(sc, b, b.Op, string(b.Op))
	if err != nil {
		return err
	}
	if affected == 0 { // duplicated or hanging request
		return txbarrier.ErrDuplicationOrSuspension
	}

	if b.Op == txbarrier.Cancel {
		affected, err = insertDB(sc, b, txbarrier.Try, string(b.Op))
		if err != nil {
			return err
		}
		if affected > 0 { // empty compensation
			return txbarrier.ErrEmptyCompensation
		}
	}

	return nil
}

func insertDB(sc mongo.SessionContext, b *txbarrier.Barrier, op txbarrier.Operation, reason string) (int64, error) {
	tmp := strings.Split(DefaultDBCollection, ".")
	if len(tmp) != 2 {
		return 0, fmt.Errorf("invalid db collection name `%s`", DefaultDBCollection)
	}

	_, err := sc.Client().Database(tmp[0]).Collection(tmp[1]).InsertOne(sc, bson.D{
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

package mgobarrier

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/magic-lib/go-plat-utils/db/txbarrier"
)

var testDuplicateErrRsp = mtest.CreateWriteErrorsResponse(mtest.WriteError{
	Index:   1,
	Code:    11000,
	Message: "duplicate key error",
})

func Test_DoWithClient(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("duplicated or hanging request", func(t *mtest.T) {
		DefaultDBCollection = t.Coll.Name() + ".txbarrier"
		t.AddMockResponses(testDuplicateErrRsp)           // for insert into db
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for session.AbortTransaction

		ctx := txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Try,
		})
		cnt := 0
		err := DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			cnt++
			return nil
		})
		require.Equal(t, txbarrier.ErrDuplicationOrSuspension, err)
		require.Zero(t, cnt)
	})
	mt.Run("empty compensation", func(t *mtest.T) {
		DefaultDBCollection = t.Coll.Name() + ".txbarrier"
		t.AddMockResponses(mtest.CreateSuccessResponse()) // first rsp for Cancel
		t.AddMockResponses(mtest.CreateSuccessResponse()) // second rsp for insertion of Try
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the session.CommitTransaction

		ctx := txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Cancel,
		})
		cnt := 0
		err := DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			cnt++
			return nil
		})
		require.Equal(t, txbarrier.ErrEmptyCompensation, err)
		require.Zero(t, cnt)
	})
	mt.Run("normal success", func(t *mtest.T) {
		DefaultDBCollection = t.Coll.Name() + ".txbarrier"
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for insertion of Try
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the "Try" session.CommitTransaction
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for insertion of Confirm
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the "Confirm" session.CommitTransaction

		ctx := txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Try,
		})
		cnt := 0
		err := DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			cnt++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, cnt)

		ctx = txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Confirm,
		})
		err = DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			cnt++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 2, cnt)
	})
	mt.Run("normal cancel", func(t *mtest.T) {
		DefaultDBCollection = t.Coll.Name() + ".txbarrier"
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for insertion of Try
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the "Try" session.CommitTransaction
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for insertion of Cancel
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the "Cancel" session.CommitTransaction

		ctx := txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Try,
		})
		cnt := 0
		err := DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			cnt++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, cnt)

		ctx = txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Confirm,
		})
		err = DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			cnt--
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 0, cnt)
	})
	mt.Run("normal failed", func(t *mtest.T) {
		testBizErr := fmt.Errorf("business error")
		DefaultDBCollection = t.Coll.Name() + ".txbarrier"
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for insertion of Try
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the "Try" session.AbortTransaction
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for insertion of Confirm
		t.AddMockResponses(mtest.CreateSuccessResponse()) // for the "Confirm" session.AbortTransaction

		// Try: business error
		ctx := txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "1", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Try,
		})
		err := DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			return testBizErr
		})
		require.Contains(t, err.Error(), "business error")

		// Confirm: business error
		ctx = txbarrier.NewCtxWithBarrier(context.Background(), &txbarrier.Barrier{
			XID: "2", BranchID: "2", TransTyp: "tcc", Op: txbarrier.Confirm,
		})
		err = DoWithClient(ctx, t.Client, func(sc mongo.SessionContext) error {
			return testBizErr
		})
		require.Contains(t, err.Error(), "business error")
	})
}

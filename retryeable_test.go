package exchanges

import (
	"context"
	"errors"
	"testing"

	"github.com/aulaleslie/trade-exchanges/utils"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPlaceBuyOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ex := NewMockExchange(ctrl)

	re := &RetryeableExchange{Target: ex}

	gomock.InOrder(
		ex.EXPECT().
			PlaceBuyOrder(gomock.Any(), false, "S", utils.FromUint(5), utils.FromUint(6), "id1").
			Return("", errors.New("err")).Times(1),
		ex.EXPECT().
			PlaceBuyOrder(gomock.Any(), true, "S", utils.FromUint(5), utils.FromUint(6), "id1").
			Return("id2", nil).Times(1),
	)
	id2, err := re.PlaceBuyOrder(context.TODO(), false, "S", utils.FromUint(5), utils.FromUint(6), "id1")
	assert.Equal(t, "id2", id2)
	assert.NoError(t, err)
}

func TestPlaceSellOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ex := NewMockExchange(ctrl)

	re := &RetryeableExchange{Target: ex}

	gomock.InOrder(
		ex.EXPECT().
			PlaceSellOrder(gomock.Any(), true, "S", utils.FromUint(5), utils.FromUint(6), "id1").
			Return("", errors.New("err")).Times(1),
		ex.EXPECT().
			PlaceSellOrder(gomock.Any(), true, "S", utils.FromUint(5), utils.FromUint(6), "id1").
			Return("id2", nil).Times(1),
	)

	id2, err := re.PlaceSellOrder(context.TODO(), true, "S", utils.FromUint(5), utils.FromUint(6), "id1")
	assert.Equal(t, "id2", id2)
	assert.NoError(t, err)
}

func TestGetOrderInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ex := NewMockExchange(ctrl)

	re := &RetryeableExchange{Target: ex}

	expected := OrderInfo{ClientOrderID: &[]string{"the_coid"}[0], ID: "the_id", Status: ExpiredOST}
	ex.EXPECT().GetOrderInfo(gomock.Any(), "S", "the_id", nil).Return(expected, nil)
	actual, err := re.GetOrderInfo(context.TODO(), "S", "the_id", nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGetOrderInfoByClientOrderID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ex := NewMockExchange(ctrl)

	re := &RetryeableExchange{Target: ex}

	expected := OrderInfo{ClientOrderID: &[]string{"the_coid"}[0], ID: "the_id", Status: ExpiredOST}
	ex.EXPECT().GetOrderInfoByClientOrderID(gomock.Any(), "S", "the_coid", nil).Return(expected, nil)
	actual, err := re.GetOrderInfoByClientOrderID(context.Background(), "S", "the_coid", nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestLastErrorForCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ex := NewMockExchange(ctrl)

	re := &RetryeableExchange{Target: ex}

	errFirst := errors.New("err1")
	errLast := errors.New("err2")
	gomock.InOrder(
		ex.EXPECT().CancelOrder(gomock.Any(), "S", "id1").Return(errFirst).Times(4),
		ex.EXPECT().CancelOrder(gomock.Any(), "S", "id1").Return(errLast).Times(1),
	)

	err := re.CancelOrder(context.Background(), "S", "id1")
	assert.Error(t, err)
	assert.Equal(t, errLast, err)
}

func TestLastError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ex := NewMockExchange(ctrl)

	re := &RetryeableExchange{Target: ex}

	errFirst := errors.New("err1")
	errLast := errors.New("err2")
	gomock.InOrder(
		ex.EXPECT().GetOrderInfo(gomock.Any(), "S", "id1", nil).Return(OrderInfo{}, errFirst).Times(2),
		ex.EXPECT().GetOrderInfo(gomock.Any(), "S", "id1", nil).Return(OrderInfo{}, errLast).Times(1),
	)

	_, err := re.GetOrderInfo(context.Background(), "S", "id1", nil)
	assert.Error(t, err)
	assert.Equal(t, errLast, err)
}

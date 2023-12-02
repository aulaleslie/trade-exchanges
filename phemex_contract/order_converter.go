package phemex_contract

import (
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/pkg/errors"
)

// Can be used only in case if `bizError` field can't be in respond order
func convertOrderStatus(phemexOrderStatus string) (exchanges.OrderStatusType, error) {
	switch phemexOrderStatus {
	case "Untriggered": // Conditional order waiting to be triggered
		return exchanges.UnknownOST, nil
	case "Triggered": // Conditional order being triggered
		return exchanges.UnknownOST, nil
	case "Rejected":
		return exchanges.RejectedOST, nil
	case "New":
		return exchanges.NewOST, nil
	case "PartiallyFilled":
		return exchanges.PartiallyFilledOST, nil
	case "Filled":
		return exchanges.FilledOST, nil
	case "Canceled":
		return exchanges.CanceledOST, nil
	case "Created": // https://github.com/phemex/phemex-api-docs/blob/master/Public-Contract-API-en.md#place-order
		return exchanges.NewOST, nil
	default:
		return exchanges.UnknownOST, errors.Errorf(
			"unknown order status = %s", phemexOrderStatus)
	}
}

const OrderNotFoundCode = 10002 // OM_ORDER_NOT_FOUND

func checkBizError(phemexOrder *krisa_phemex_fork.OrderResponse) (exchanges.OrderStatusType, error) {
	if phemexOrder.BizError == OrderNotFoundCode {
		return exchanges.UnknownOST, PhemexOrderClearedError
	}
	if isOrderRejectedCode(int64(phemexOrder.BizError)) {
		return exchanges.RejectedOST, nil
	}

	if phemexOrder.BizError != 0 {
		return exchanges.UnknownOST, errors.Errorf("invalid BizError=%d, orderID=%s",
			phemexOrder.BizError, phemexOrder.OrderID)
	}
	return exchanges.UnknownOST, nil
}

func convertOrder(phemexOrder *krisa_phemex_fork.OrderResponse) (exchanges.OrderInfo, error) {
	bizStatus, err := checkBizError(phemexOrder)
	if err != nil {
		return exchanges.OrderInfo{}, errors.Wrap(err, "invalid bizError")
	}

	status, err := convertOrderStatus(phemexOrder.OrdStatus)
	if err != nil {
		return exchanges.OrderInfo{}, errors.Wrap(err, "can't convert status")
	}

	switch {
	case bizStatus == exchanges.UnknownOST:
		// ok
	case bizStatus == status:
		// ok
	case bizStatus != status:
		return exchanges.OrderInfo{}, errors.Errorf(
			"unknown statuses (bizStatus=%v, status=%v)", bizStatus, status)
	}

	return exchanges.OrderInfo{
		ID:            phemexOrder.OrderID,
		ClientOrderID: &phemexOrder.ClOrdID,
		Status:        status,
	}, nil
}

func isOrderRejectedCode(code int64) bool {
	_, found := orderRejectedCodes[code]
	return found
}

var orderRejectedCodes map[int64]struct{} = map[int64]struct{}{
	/// 19999 : BadRequest          # REQUEST_IS_DUPLICATED                   # Duplicated request ID
	/// 10001 : DuplicateOrderId    # OM_DUPLICATE_ORDERID                    # Duplicated order ID
	/// 10002 : OrderNotFound       # OM_ORDER_NOT_FOUND                      # Cannot find order ID
	/// 10003 : CancelPending       # OM_ORDER_PENDING_CANCEL                 # Cannot cancel while order is already in pending cancel status
	/// 10004 : CancelPending       # OM_ORDER_PENDING_REPLACE                # Cannot cancel while order is already in pending cancel status
	/// 10005 : CancelPending       # OM_ORDER_PENDING                        # Cannot cancel while order is already in pending cancel status
	11001: {}, // InsufficientFunds # TE_NO_ENOUGH_AVAILABLE_BALANCE          # Insufficient available balance
	11002: {}, // InvalidOrder      # TE_INVALID_RISK_LIMIT                   # Invalid risk limit value
	11003: {}, // InsufficientFunds # TE_NO_ENOUGH_BALANCE_FOR_NEW_RISK_LIMIT # Insufficient available balance
	11004: {}, // InvalidOrder      # TE_INVALID_LEVERAGE                     # invalid input or new leverage is over maximum allowed leverage
	11005: {}, // InsufficientFunds # TE_NO_ENOUGH_BALANCE_FOR_NEW_LEVERAGE   # Insufficient available balance
	11006: {}, // ExchangeError     # TE_CANNOT_CHANGE_POSITION_MARGIN_WITHOUT_POSITION      # Position size is zero. Cannot change margin
	11007: {}, // ExchangeError     # TE_CANNOT_CHANGE_POSITION_MARGIN_FOR_CROSS_MARGIN      # Cannot change margin under CrossMargin
	11008: {}, // ExchangeError     # TE_CANNOT_REMOVE_POSITION_MARGIN_MORE_THAN_ADDED       # exceeds the maximum removable Margin
	11009: {}, // ExchangeError     # TE_CANNOT_REMOVE_POSITION_MARGIN_DUE_TO_UNREALIZED_PNL # exceeds the maximum removable Margin
	11010: {}, // InsufficientFunds # TE_CANNOT_ADD_POSITION_MARGIN_DUE_TO_NO_ENOUGH_AVAILABLE_BALANCE # Insufficient available balance
	11011: {}, // InvalidOrder      # TE_REDUCE_ONLY_ABORT                     # Cannot accept reduce only order
	11012: {}, // InvalidOrder      # TE_REPLACE_TO_INVALID_QTY                # Order quantity Error
	11013: {}, // InvalidOrder      # TE_CONDITIONAL_NO_POSITION               # Position size is zero. Cannot determine conditional order's quantity'
	11014: {}, // InvalidOrder      # TE_CONDITIONAL_CLOSE_POSITION_WRONG_SIDE # Close position conditional order has the same side
	11015: {}, // InvalidOrder      # TE_CONDITIONAL_TRIGGERED_OR_CANCELED     # -
	/// 11016 : BadRequest          # TE_ADL_NOT_TRADING_REQUESTED_ACCOUNT     # Request is routed to the wrong trading engine
	/// 11017 : ExchangeError       # TE_ADL_CANNOT_FIND_POSITION              # Cannot find requested position on current account
	/// 11018 : ExchangeError       # TE_NO_NEED_TO_SETTLE_FUNDING             # The current account does not need to pay a funding fee
	/// 11019 : ExchangeError       # TE_FUNDING_ALREADY_SETTLED               # The current account already pays the funding fee
	/// 11020 : ExchangeError       # TE_CANNOT_TRANSFER_OUT_DUE_TO_BONUS      # Withdraw to wallet needs to remove all remaining bonus. However if bonus is used by position or order cost, withdraw fails.
	/// 11021 : ExchangeError       # TE_INVALID_BONOUS_AMOUNT                 # #  Grpc command cannot be negative number Invalid bonus amount
	/// 11022 : AccountSuspended    # TE_REJECT_DUE_TO_BANNED                  # Account is banned
	/// 11023 : ExchangeError       # TE_REJECT_DUE_TO_IN_PROCESS_OF_LIQ       # Account is in the process of liquidation
	/// 11024 : ExchangeError       # TE_REJECT_DUE_TO_IN_PROCESS_OF_ADL       # Account is in the process of auto-deleverage
	/// 11025 : BadRequest          # TE_ROUTE_ERROR                 # Request is routed to the wrong trading engine
	/// 11026 : ExchangeError       # TE_UID_ACCOUNT_MISMATCH        # -
	/// 11027 : BadSymbol           # TE_SYMBOL_INVALID              # Invalid number ID or name
	/// 11028 : BadSymbol           # TE_CURRENCY_INVALID            # Invalid currency ID or name
	/// 11029 : ExchangeError       # TE_ACTION_INVALID              # Unrecognized request type
	/// 11030 : ExchangeError       # TE_ACTION_BY_INVALID           # -
	/// 11031 : DDoSProtection      # TE_SO_NUM_EXCEEDS              # Number of total conditional orders exceeds the max limit
	11032: {}, // DDoSProtection    # TE_AO_NUM_EXCEEDS              # Number of total active orders exceeds the max limit TODO: ask about this limit
	/// 11033 : DuplicateOrderId    # TE_ORDER_ID_DUPLICATE          # Duplicated order ID
	/// 11034 : InvalidOrder        # TE_SIDE_INVALID                # Invalid side
	/// 11035 : InvalidOrder        # TE_ORD_TYPE_INVALID            # Invalid OrderType
	/// 11036 : InvalidOrder        # TE_TIME_IN_FORCE_INVALID       # Invalid TimeInForce
	/// 11037 : InvalidOrder        # TE_EXEC_INST_INVALID           # Invalid ExecType
	/// 11038 : InvalidOrder        # TE_TRIGGER_INVALID             # Invalid trigger type
	/// 11039 : InvalidOrder        # TE_STOP_DIRECTION_INVALID      # Invalid stop direction type
	/// 11040 : InvalidOrder        # TE_NO_MARK_PRICE               # Cannot get valid mark price to create conditional order
	/// 11041 : InvalidOrder        # TE_NO_INDEX_PRICE              # Cannot get valid index price to create conditional order
	/// 11042 : InvalidOrder        # TE_NO_LAST_PRICE               # Cannot get valid last market price to create conditional order
	/// 11043 : InvalidOrder        # TE_RISING_TRIGGER_DIRECTLY     # Conditional order would be triggered immediately
	/// 11044 : InvalidOrder        # TE_FALLING_TRIGGER_DIRECTLY    # Conditional order would be triggered immediately
	/// 11045 : InvalidOrder        # TE_TRIGGER_PRICE_TOO_LARGE     # Conditional order trigger price is too high
	/// 11046 : InvalidOrder        # TE_TRIGGER_PRICE_TOO_SMALL     # Conditional order trigger price is too low
	/// 11047 : InvalidOrder        # TE_BUY_TP_SHOULD_GT_BASE       # TakeProfile BUY conditional order trigger price needs to be greater than reference price
	/// 11048 : InvalidOrder        # TE_BUY_SL_SHOULD_LT_BASE       # StopLoss BUY condition order price needs to be less than the reference price
	/// 11049 : InvalidOrder        # TE_BUY_SL_SHOULD_GT_LIQ        # StopLoss BUY condition order price needs to be greater than liquidation price or it will not trigger
	/// 11050 : InvalidOrder        # TE_SELL_TP_SHOULD_LT_BASE      # TakeProfile SELL conditional order trigger price needs to be less than reference price
	/// 11051 : InvalidOrder        # TE_SELL_SL_SHOULD_LT_LIQ       # StopLoss SELL condition order price needs to be less than liquidation price or it will not trigger
	/// 11052 : InvalidOrder        # TE_SELL_SL_SHOULD_GT_BASE      # StopLoss SELL condition order price needs to be greater than the reference price
	11053: {}, // InvalidOrder      # TE_PRICE_TOO_LARGE             # -
	11054: {}, // InvalidOrder      # TE_PRICE_WORSE_THAN_BANKRUPT   # Order price cannot be more aggressive than bankrupt price if self order has instruction to close a position
	11055: {}, // InvalidOrder      # TE_PRICE_TOO_SMALL             # Order price is too low
	11056: {}, // InvalidOrder      # TE_QTY_TOO_LARGE               # Order quantity is too large
	11057: {}, // InvalidOrder      # TE_QTY_NOT_MATCH_REDUCE_ONLY   # Does not allow ReduceOnly order without position
	11058: {}, // InvalidOrder      # TE_QTY_TOO_SMALL               # Order quantity is too small
	11059: {}, // InvalidOrder      # TE_TP_SL_QTY_NOT_MATCH_POS     # Position size is zero. Cannot accept any TakeProfit or StopLoss order
	11060: {}, // InvalidOrder      # TE_SIDE_NOT_CLOSE_POS          # TakeProfit or StopLoss order has wrong side. Cannot close position
	/// 11061 : CancelPending       # TE_ORD_ALREADY_PENDING_CANCEL  # Repeated cancel request
	/// 11062 : InvalidOrder        # TE_ORD_ALREADY_CANCELED        # Order is already canceled
	/// 11063 : InvalidOrder        # TE_ORD_STATUS_CANNOT_CANCEL    # Order is not able to be canceled under current status
	/// 11064 : InvalidOrder        # TE_ORD_ALREADY_PENDING_REPLACE # Replace request is rejected because order is already in pending replace status
	/// 11065 : InvalidOrder        # TE_ORD_REPLACE_NOT_MODIFIED    # Replace request does not modify any parameters of the order
	/// 11066 : InvalidOrder        # TE_ORD_STATUS_CANNOT_REPLACE   # Order is not able to be replaced under current status
	/// 11067 : InvalidOrder        # TE_CANNOT_REPLACE_PRICE        # Market conditional order cannot change price
	/// 11068 : InvalidOrder        # TE_CANNOT_REPLACE_QTY          # Condtional order for closing position cannot change order quantity, since the order quantity is determined by position size already
	/// 11069 : ExchangeError       # TE_ACCOUNT_NOT_IN_RANGE        # The account ID in the request is not valid or is not in the range of the current process
	/// 11070 : BadSymbol           # TE_SYMBOL_NOT_IN_RANGE         # The symbol is invalid
	/// 11071 : InvalidOrder        # TE_ORD_STATUS_CANNOT_TRIGGER   # -
	/// 11072 : InvalidOrder        # TE_TKFR_NOT_IN_RANGE           # The fee value is not valid
	/// 11073 : InvalidOrder        # TE_MKFR_NOT_IN_RANGE           # The fee value is not valid
	11074: {}, // InvalidOrder      # TE_CANNOT_ATTACH_TP_SL         # Order request cannot contain TP/SL parameters when the account already has positions
	11075: {}, // InvalidOrder      # TE_TP_TOO_LARGE                # TakeProfit price is too large
	11076: {}, // InvalidOrder      # TE_TP_TOO_SMALL                # TakeProfit price is too small
	/// 11077 : InvalidOrder        # TE_TP_TRIGGER_INVALID          # Invalid trigger type
	11078: {}, // InvalidOrder      # TE_SL_TOO_LARGE                # StopLoss price is too large
	11079: {}, // InvalidOrder      # TE_SL_TOO_SMALL                # StopLoss price is too small
	/// 11080 : InvalidOrder        # TE_SL_TRIGGER_INVALID          # Invalid trigger type
	11081: {}, // InvalidOrder      # TE_RISK_LIMIT_EXCEEDS          # Total potential position breaches current risk limit
	11082: {}, // InsufficientFunds # TE_CANNOT_COVER_ESTIMATE_ORDER_LOSS    # The remaining balance cannot cover the potential unrealized PnL for self new order
	/// 11083 : InvalidOrder        # TE_TAKE_PROFIT_ORDER_DUPLICATED        # TakeProfit order already exists
	/// 11084 : InvalidOrder        # TE_STOP_LOSS_ORDER_DUPLICATED          # StopLoss order already exists
	/// 11085 : DuplicateOrderId    # TE_CL_ORD_ID_DUPLICATE                 # ClOrdId is duplicated
	/// 11086 : InvalidOrder        # TE_PEG_PRICE_TYPE_INVALID              # PegPriceType is invalid
	11087: {}, // InvalidOrder      # TE_BUY_TS_SHOULD_LT_BASE               # The trailing order's StopPrice should be less than the current last price
	11088: {}, // InvalidOrder      # TE_BUY_TS_SHOULD_GT_LIQ                # The traling order's StopPrice should be greater than the current liquidation price
	11089: {}, // InvalidOrder      # TE_SELL_TS_SHOULD_LT_LIQ               # The traling order's StopPrice should be greater than the current last price
	11090: {}, // InvalidOrder      # TE_SELL_TS_SHOULD_GT_BASE              # The traling order's StopPrice should be less than the current liquidation price
	11091: {}, // InvalidOrder      # TE_BUY_REVERT_VALUE_SHOULD_LT_ZERO     # The PegOffset should be less than zero
	11092: {}, // InvalidOrder      # TE_SELL_REVERT_VALUE_SHOULD_GT_ZERO    # The PegOffset should be greater than zero
	11093: {}, // InvalidOrder      # TE_BUY_TTP_SHOULD_ACTIVATE_ABOVE_BASE  # The activation price should be greater than the current last price
	11094: {}, // InvalidOrder      # TE_SELL_TTP_SHOULD_ACTIVATE_BELOW_BASE # The activation price should be less than the current last price
	/// 11095 : InvalidOrder        # TE_TRAILING_ORDER_DUPLICATED           # A trailing order exists already
	11096: {}, // InvalidOrder      # TE_CLOSE_ORDER_CANNOT_ATTACH_TP_SL     # An order to close position cannot have trailing instruction
	/// 11097 : BadRequest          # TE_CANNOT_FIND_WALLET_OF_THIS_CURRENCY # This crypto is not supported
	/// 11098 : BadRequest          # TE_WALLET_INVALID_ACTION               # Invalid action on wallet
	/// 11099 : ExchangeError       # TE_WALLET_VID_UNMATCHED                # Wallet operation request has a wrong wallet vid
	11100: {}, // InsufficientFunds # TE_WALLET_INSUFFICIENT_BALANCE         # Wallet has insufficient balance
	11101: {}, // InsufficientFunds # TE_WALLET_INSUFFICIENT_LOCKED_BALANCE  # Locked balance in wallet is not enough for unlock/withdraw request
	11102: {}, // BadRequest        # TE_WALLET_INVALID_DEPOSIT_AMOUNT       # Deposit amount must be greater than zero
	/// 11103 : BadRequest          # TE_WALLET_INVALID_WITHDRAW_AMOUNT      # Withdraw amount must be less than zero
	/// 11104 : BadRequest          # TE_WALLET_REACHED_MAX_AMOUNT           # Deposit makes wallet exceed max amount allowed
	/// 11105 : InsufficientFunds   # TE_PLACE_ORDER_INSUFFICIENT_BASE_BALANCE  # Insufficient funds in base wallet
	11106: {}, // InsufficientFunds # TE_PLACE_ORDER_INSUFFICIENT_QUOTE_BALANCE # Insufficient funds in quote wallet
	11107: {}, // ExchangeError     # TE_CANNOT_CONNECT_TO_REQUEST_SEQ          # TradingEngine failed to connect with CrossEngine
	/// 11108 : InvalidOrder        # TE_CANNOT_REPLACE_OR_CANCEL_MARKET_ORDER  # Cannot replace/amend market order
	/// 11109 : InvalidOrder        # TE_CANNOT_REPLACE_OR_CANCEL_IOC_ORDER     # Cannot replace/amend ImmediateOrCancel order
	/// 11110 : InvalidOrder        # TE_CANNOT_REPLACE_OR_CANCEL_FOK_ORDER     # Cannot replace/amend FillOrKill order
	/// 11111 : InvalidOrder        # TE_MISSING_ORDER_ID      # OrderId is missing
	/// 11112 : InvalidOrder        # TE_QTY_TYPE_INVALID      # QtyType is invalid
	/// 11113 : BadRequest          # TE_USER_ID_INVALID       # UserId is invalid
	11114: {}, // InvalidOrder      # TE_ORDER_VALUE_TOO_LARGE # Order value is too large
	11115: {}, // InvalidOrder      # TE_ORDER_VALUE_TOO_SMALL # Order value is too small

	/// 11116 | TE_BO_NUM_EXCEEDS                         | the total count of brakcet orders should equal or less than 5
	/// 11117 | TE_BO_CANNOT_HAVE_BO_WITH_DIFF_SIDE       | all bracket orders should have the same Side.
	11118: {}, // TE_BO_TP_PRICE_INVALID                  | bracker order take profit price is invalid
	11119: {}, // TE_BO_SL_PRICE_INVALID                  | bracker order stop loss price is invalid
	11120: {}, // TE_BO_SL_TRIGGER_PRICE_INVALID          | bracker order stop loss trigger price is invalid
	/// 11121 | TE_BO_CANNOT_REPLACE                      | cannot replace bracket order.
	11122: {}, // TE_BO_BOTP_STATUS_INVALID               | bracket take profit order status is invalid
	11123: {}, // TE_BO_CANNOT_PLACE_BOTP_OR_BOSL_ORDER   | cannot place bracket take profit order
	11124: {}, // TE_BO_CANNOT_REPLACE_BOTP_OR_BOSL_ORDER | cannot place bracket stop loss order
	/// 11125 | TE_BO_CANNOT_CANCEL_BOTP_OR_BOSL_ORDER    | cannot cancel bracket sl/tp order
	/// 11126 | TE_BO_DONOT_SUPPORT_API                   | doesn't support bracket order via API
	/// 11128 | TE_BO_INVALID_EXECINST                    | ExecInst value is invalid
	/// 11129 | TE_BO_MUST_BE_SAME_SIDE_AS_POS            | bracket order should have the same side as position's side
	/// 11130 | TE_BO_WRONG_SL_TRIGGER_TYPE               | bracket stop loss order trigger type is invalid
	/// 11131 | TE_BO_WRONG_TP_TRIGGER_TYPE               | bracket take profit order trigger type is invalid
	/// 11132 | TE_BO_ABORT_BOSL_DUE_BOTP_CREATE_FAILED   | cancel bracket stop loss order due failed to create take profit order.
	/// 11133 | TE_BO_ABORT_BOSL_DUE_BOPO_CANCELED        | cancel bracket stop loss order due main order canceled.
	/// 11134 | TE_BO_ABORT_BOTP_DUE_BOPO_CANCELED        | cancel bracket take profit order due main order canceled.

	/// 19999 # BadRequest # REQUEST_IS_DUPLICATED # Duplicated request ID
}

package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ivaylopivanov/chaincode-samples/multi-currency-wallet/numbers"
	"github.com/ivaylopivanov/chaincode-samples/multi-currency-wallet/transactions"
)

func update(fn string, stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 4 {
		return shim.Error("Not enough arguments")
	}

	method := args[0]
	user := args[1]
	currency := args[2]
	value := args[3]

	current, err := stub.GetState(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get current balance error: %s", err))
	}

	state, err := newState(current)
	if err != nil {
		return shim.Error(fmt.Sprintf("Parse current balance error: %s", err))
	}

	newBalance, err := calculateNewBalance(fn, value, state.get(currency))
	if err != nil {
		return shim.Error(fmt.Sprintf("Update balance error: %s", err))
	}

	state.set(currency, newBalance)

	b, err := state.toJSON()
	if err != nil {
		return shim.Error(fmt.Sprintf("toJSON error: %s", err))
	}

	if err := stub.PutState(user, b); err != nil {
		return shim.Error(fmt.Sprintf("Save balance error: %s", err))
	}

	t, err := transactions.Save(stub, method, user, currency, value)

	if err != nil {
		return shim.Error(fmt.Sprintf("Save transaction error: %s", err))
	}

	return shim.Success(t)
}

func calculateNewBalance(fn string, change string, current float64) (float64, error) {
	value, err := numbers.StringToFloat64(change)
	if err != nil {
		return 0, err
	}

	if value < 0 {
		// we cannot have negative values which may produce an undesirable result
		// eg: 20 - -10 or 20 + - 10
		return 0, errors.New("Passing a negative value")
	}

	if current == 0 {
		// the user has no balance yet - eg new user
		return value, nil
	}

	if fn == "deposit" {
		return current + value, nil
	}
	// otherwhise it's a withdrawal
	newBalance := current - value
	if newBalance < 0 {
		return 0, errors.New("insufficient funds")
	}

	return newBalance, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type SmartContract struct {
}

type CounterNO struct {
	Counter int `json:"counter"`
}

type User struct {
	UserID      string `json:"UserID"`
	UserName    string `json:"UserName"`
	UserBalance int64  `json:"UserBalance"`
}
type Product struct {
	ProductID    string `json:"ProductID"`
	ProductName  string `json:"ProductName"`
	ProductPrice int64  `json:"ProductPrice"`
	Owner        User   `json:"Owner"`
}
type Transaction struct {
	TransactionID    string  `json:"TransactionID"`
	OldOwner         User    `json:"OldOwner"`
	NewOwner         User    `json:"NewOwner"`
	Product          Product `json:"Product"`
	TransactionPrice int64   `json:"TransactionPrice"`
	TransactionDate  string  `json:"TransactionDate"`
}

// =================================================================================== // Main // ===================================================================================

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode // ===========================

func (t *SmartContract) Init(APIstub shim.ChaincodeStubInterface) pb.Response {

	// Initializing Product Counter
	ProductCounterBytes, _ := APIstub.GetState("ProductCounterNO")
	if ProductCounterBytes == nil {
		var ProductCounter = CounterNO{Counter: 0}
		ProductCounterBytes, _ := json.Marshal(ProductCounter)
		err := APIstub.PutState("ProductCounterNO", ProductCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Product Counter"))
		}
	}
	// Initializing Order Counter
	TransactionCounterBytes, _ := APIstub.GetState("TransactionCounterNO")
	if TransactionCounterBytes == nil {
		var TransactionCounter = CounterNO{Counter: 0}
		TransactionCounterBytes, _ := json.Marshal(TransactionCounter)
		err := APIstub.PutState("TransactionCounterNO", TransactionCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Transaction Counter"))
		}
	}
	// Initializing User Counter
	UserCounterBytes, _ := APIstub.GetState("UserCounterNO")
	if UserCounterBytes == nil {
		var UserCounter = CounterNO{Counter: 0}
		UserCounterBytes, _ := json.Marshal(UserCounter)
		err := APIstub.PutState("UserCounterNO", UserCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate User Counter"))
		}
	}

	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations // ========================================

func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initLedger" {
		//init ledger
		return t.initLedger(stub, args)
	} else if function == "createUser" { //username, balance
		//create a new user
		return t.createUser(stub, args)
	} else if function == "createProduct" { //product name, price, owner
		//create a new product
		return t.createProduct(stub, args)
	} else if function == "updateProduct" { //old owner id, new owner id, product id
		// update a product
		return t.updateProduct(stub, args)
	} else if function == "orderProduct" {
		// order a product
		return t.orderProduct(stub, args)
	} else if function == "queryAsset" {
		// query any using asset-id
		return t.queryAsset(stub, args)
	} else if function == "queryAll" {
		// query all assests of a type
		return t.queryAll(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)
	//error
	return shim.Error("Received unknown function invocation")
}

// Private function

//getCounter to the latest value of the counter based on the Asset Type provided as input parameter
func getCounter(APIstub shim.ChaincodeStubInterface, AssetType string) int {
	counterAsBytes, _ := APIstub.GetState(AssetType)
	counterAsset := CounterNO{}

	json.Unmarshal(counterAsBytes, &counterAsset)
	fmt.Sprintf("Counter Current Value %d of Asset Type %s", counterAsset.Counter, AssetType)

	return counterAsset.Counter
}

//incrementCounter to the increase value of the counter based on the Asset Type provided as input parameter by 1
func incrementCounter(APIstub shim.ChaincodeStubInterface, AssetType string) int {
	counterAsBytes, _ := APIstub.GetState(AssetType)
	counterAsset := CounterNO{}

	json.Unmarshal(counterAsBytes, &counterAsset)
	counterAsset.Counter++
	counterAsBytes, _ = json.Marshal(counterAsset)

	err := APIstub.PutState(AssetType, counterAsBytes)
	if err != nil {

		fmt.Sprintf("Failed to Increment Counter")

	}

	fmt.Println("Success in incrementing counter  %v", counterAsset)

	return counterAsset.Counter
}

// GetTxTimestampChannel Function gets the Transaction time when the chain code was executed it remains same on all the peers where chaincode executes
func (t *SmartContract) GetTxTimestampChannel(APIstub shim.ChaincodeStubInterface) (string, error) {
	txTimeAsPtr, err := APIstub.GetTxTimestamp()
	if err != nil {
		fmt.Printf("Returning error in TimeStamp \n")
		return "Error", err
	}
	fmt.Printf("\t returned value from APIstub: %v\n", txTimeAsPtr)
	timeStr := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos)).String()

	return timeStr, nil
}

func (t *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	// seed admin
	entityUser := User{UserID: "admin", UserName: "admin", UserBalance: 1000000}
	entityUserAsBytes, errMarshal := json.Marshal(entityUser)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in user: %s", errMarshal))
	}

	errPut := APIstub.PutState(entityUser.UserID, entityUserAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Entity Asset: %s", entityUser.UserID))
	}

	fmt.Println("Added", entityUser)

	return shim.Success(nil)
}

//create user
//username, balance
func (t *SmartContract) createUser(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	balance, err := strconv.ParseInt(args[1], 10, 0)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(args) != 2 { //name, balance , id auto generate
		return shim.Error("args are not correct")
	}
	userCounter := getCounter(APIstub, "UserCounterNO")
	userCounter++

	user := User{UserID: "User" + strconv.Itoa(userCounter), UserName: args[0], UserBalance: balance}

	userBytes, errMarshal := json.Marshal(user)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}

	errPut := APIstub.PutState(user.UserID, userBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to register user: %s", user.UserID))
	}

	//TO Increment the User Counter
	incrementCounter(APIstub, "UserCounterNO")

	fmt.Println("User register successfully %v", user)
	return shim.Success(userBytes)

}

func (t *SmartContract) createProduct(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	//To check number of arguments are 3
	//productName, product price , user id
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, Required 3 arguments")
	}

	//Price conversion - Error handeling
	price, errPrice := strconv.ParseInt(args[1], 10, 0)
	if errPrice != nil {
		return shim.Error(fmt.Sprintf("Failed to Convert Price: %s", errPrice))
	}
	userBytes, _ := APIstub.GetState(args[2])

	if userBytes == nil {
		return shim.Error("Cannot Find User")
	}
	// get user details from the stub ie. Chaincode stub in network using the user id passed

	user := User{}

	// unmarsahlling product the data from API
	json.Unmarshal(userBytes, &user)
	productCounter := getCounter(APIstub, "ProductCounterNO")
	productCounter++

	var product = Product{ProductID: "Product" + strconv.Itoa(productCounter),
		ProductName: args[0], ProductPrice: price, Owner: user}

	productBytes, errMarshal := json.Marshal(product)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}

	errPut := APIstub.PutState(product.ProductID, productBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Product Asset: %s", product.ProductID))
	}

	//TO Increment the Product Counter
	incrementCounter(APIstub, "ProductCounterNO")

	fmt.Println("Success in creating Product Asset %v", product)

	return shim.Success(productBytes)
}

// function to update the product name and price
// Input params : product id , product name , preduct price , user id
func (t *SmartContract) updateProduct(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	// parameter length check
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments, Required 4")
	}

	// get user details from the stub ie. Chaincode stub in network using the user id passed
	userBytes, _ := APIstub.GetState(args[3])

	if userBytes == nil {
		return shim.Error("Cannot Find User")
	}

	user := User{}

	// unmarsahlling product the data from API
	json.Unmarshal(userBytes, &user)

	// get product details from the stub ie. Chaincode stub in network using the product id passed
	productBytes, _ := APIstub.GetState(args[0])
	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}
	product := Product{}

	// unmarsahlling product the data from API
	json.Unmarshal(productBytes, &product)

	//Price conversion - Error handeling
	price, errPrice := strconv.ParseInt(args[2], 10, 0)
	if errPrice != nil {
		return shim.Error(fmt.Sprintf("Failed to Convert Price: %s", errPrice))
	}

	// Updating the product values withe the new values
	product.ProductName = args[2] // product name from UI for the update
	product.ProductPrice = price  // product value from UI for the update

	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := APIstub.PutState(product.ProductID, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to Sell To Cosumer : %s", product.ProductID))
	}

	fmt.Println("Success in updating Product %v ", product.ProductID)

	return shim.Success(updatedProductAsBytes)
}

//old owner id, new owner id, product id
func (t *SmartContract) orderProduct(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	// parameter length check
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, Required 4")
	}

	oldOwnerBytes, _ := APIstub.GetState(args[0])
	newOwnerBytes, _ := APIstub.GetState(args[1])
	if oldOwnerBytes == nil || newOwnerBytes == nil {
		return shim.Error("Cannot Find Consumer")
	}

	oldOwner := User{}
	newOwner := User{}

	// unmarsahlling product the data from API
	json.Unmarshal(oldOwnerBytes, &oldOwner)
	json.Unmarshal(newOwnerBytes, &newOwner)

	productBytes, _ := APIstub.GetState(args[2])
	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}
	product := Product{}

	// unmarsahlling product the data from API
	json.Unmarshal(productBytes, &product)

	txPrice := product.ProductPrice
	oldOwner.UserBalance += txPrice
	newOwner.UserBalance -= txPrice

	product.Owner = newOwner

	oldOwnerBytes, _ = json.Marshal(oldOwner)
	newOwnerBytes, _ = json.Marshal(newOwner)
	productBytes, _ = json.Marshal(product)
	fmt.Println("oldowner refresh : ", string(oldOwnerBytes))
	fmt.Println("newowner refresh : ", string(newOwnerBytes))
	APIstub.PutState(oldOwner.UserID, oldOwnerBytes)
	APIstub.PutState(newOwner.UserID, newOwnerBytes)
	APIstub.PutState(product.ProductID, productBytes)
	transactionCounter := getCounter(APIstub, "TransactionCounterNO")
	transactionCounter++

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(APIstub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	transaction := Transaction{}

	transaction.TransactionID = "Transaction" + strconv.Itoa(transactionCounter)
	transaction.OldOwner = oldOwner
	transaction.NewOwner = newOwner
	transaction.Product = product
	transaction.TransactionDate = txTimeAsPtr
	transaction.TransactionPrice = product.ProductPrice

	transactionBytes, _ := json.Marshal(transaction)
	err := APIstub.PutState(transaction.TransactionID, transactionBytes)
	if err != nil {
		shim.Error(err.Error())
	}
	incrementCounter(APIstub, "TransactionCounterNO")

	fmt.Println("Order placed successfuly %v ", transaction.TransactionID)

	return shim.Success(transactionBytes)
}

//queryAsset, id
func (t *SmartContract) queryAsset(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1 argument")
	}

	productAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(productAsBytes)
}

// query all asset of a type
func (t *SmartContract) queryAll(APIstub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, Required 1")
	}

	// parameter null check
	if len(args[0]) == 0 {
		return shim.Error("Asset Type must be provided")
	}

	assetType := args[0]
	assetCounter := getCounter(APIstub, assetType+"CounterNO")

	startKey := assetType + "1"
	endKey := assetType + strconv.Itoa(assetCounter+1)

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)

	if err != nil {

		return shim.Error(err.Error())

	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults

	var buffer bytes.Buffer

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {

		queryResponse, err := resultsIterator.Next()
		// respValue := string(queryResponse.Value)
		if err != nil {

			return shim.Error(err.Error())

		}

		// Add a comma before array members, suppress it for the first array member

		if bArrayMemberAlreadyWritten == true {

			buffer.WriteString(",")

		}

		buffer.WriteString("{\"Key\":")

		buffer.WriteString("\"")

		buffer.WriteString(queryResponse.Key)

		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")

		// Record is a JSON object, so we write as-is

		buffer.WriteString(string(queryResponse.Value))

		buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true

	}

	buffer.WriteString("]")

	fmt.Printf("- queryAllAssets:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, " ", ""); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

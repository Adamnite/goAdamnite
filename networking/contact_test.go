package networking

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/stretchr/testify/assert"
)

func TestConnectionStatusMath(t *testing.T) {
	testContacts := make([]*Contact, 10)
	conBook := NewContactBook(nil)
	for i := range testContacts {
		add := common.Address{}
		add.SetBytes(big.NewInt(int64(i)).Bytes())
		testContacts[i] = &Contact{"1.2.3.4:" + fmt.Sprint(i), add}

		if err := conBook.AddConnection(testContacts[i]); err != nil {
			t.Fatalf(err.Error())
		}
		//each test node has a delay (in nanoseconds... so these appear REALLY fast) of their index.
		conBook.connections[i].connectionAttempts = 1
		conBook.connections[i].responseTimes = []int64{int64(i)}

		assert.Equal(t, int64(i), conBook.connections[i].getAverageResponseTime(), "math to calculate base response time appears to have broken.")
	}
	assert.Equal(
		t,
		int64((len(testContacts)-1)/2),
		conBook.GetAverageConnectionResponseTime(),
		"Math for avg calculation must me off",
	)

	conBook.DropSlowestPercentage(0.1)
	assert.Equal(
		t,
		len(testContacts)-1,
		len(conBook.connections),
		"last member must have not been dropped",
	)
	assert.Equal(
		t,
		int64((len(testContacts)-2)/2),
		conBook.GetAverageConnectionResponseTime(),
		"Math for avg calculation must me off",
	)
	for i := 0; i < len(testContacts)-2; i++ {
		if !assert.LessOrEqual(
			t,
			conBook.connections[i].getAverageResponseTime(),
			conBook.connections[i+1].getAverageResponseTime(),
			"sorting must be off.",
		) {
			t.Fail()
		}

	}
	fmt.Println(conBook.GetContactList().NodeIDs)
	fmt.Println(conBook.GetContactList().ConnectionStrings)
	conBook.AddToBlacklist(testContacts[0])
	fmt.Println(conBook.GetContactList().NodeIDs)
	fmt.Println(conBook.GetContactList().ConnectionStrings)

}

func TestWhitelistGeneration(t *testing.T) {
	testContacts := make([]*Contact, 5000)
	// testContacts := make([]*Contact, 500)
	conBook := NewContactBook(nil)

	for i := range testContacts {
		add := common.Address{}
		add.SetBytes(big.NewInt(int64(i)).Bytes())
		testContacts[i] = &Contact{"1.2.3.4:" + fmt.Sprint(i), add}

		if err := conBook.AddConnection(testContacts[i]); err != nil {
			t.Fatalf(err.Error())
		}
		//each test node has a delay (in nanoseconds... so these appear REALLY fast) of their index.
		conBook.connections[i].connectionAttempts = 1
		conBook.connections[i].responseTimes = []int64{int64(i)}
	}

	// test connections. should average (eventually) to be in order by performance speed.
	whiteListLength := 40
	totalTimes := make([]int64, whiteListLength)
	var attemptCount int64 = 100000
	for i := 0; i < int(attemptCount); i++ {
		// responses = append(responses, conBook.SelectWhitelist(len(conBook.connections)))
		response := conBook.SelectWhitelist(whiteListLength)
		for x, r := range response {
			totalTimes[x] += conBook.connectionsByContact[r].getAverageResponseTime()
		}
	}

	var outOfOrderCount int64 = 0
	for i := 0; i < len(totalTimes)-1; i++ {
		if totalTimes[i]/attemptCount > totalTimes[i+1]/attemptCount {
			//you can assume some will be out of order (no matter how many times we test), i just want less than 10% out of order
			outOfOrderCount += 1
		}
	}
	assert.LessOrEqual(
		t,
		outOfOrderCount,
		int64(whiteListLength/10),
		fmt.Sprintf("%v%% of the average responses were out of order.",
			float64(outOfOrderCount)/float64(whiteListLength)*100),
	)
}

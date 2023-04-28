package networking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionStatusMath(t *testing.T) {
	// fmt.Println(MISSED_CONNECTION_TIME_PENALTY)
	testContacts := make([]*Contact, 10)
	conBook := NewContactBook()
	// testContacts[0] = &Contact{"1.2.3.4:1234", 1}
	// testContacts[1] = &Contact{"1.2.3.4:1235", 2}
	for i, _ := range testContacts {
		testContacts[i] = &Contact{"1.2.3.4:" + fmt.Sprint(i), i}
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

	fmt.Println(conBook.GetContactList().NodeIDs)
	fmt.Println(conBook.GetContactList().ConnectionStrings)
	conBook.AddToBlacklist(testContacts[1])
	fmt.Println(conBook.GetContactList().NodeIDs)
	fmt.Println(conBook.GetContactList().ConnectionStrings)
}

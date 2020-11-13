package main

import (
	"fmt"
	"math/rand"
	"time"
)

var tills []till
var customers = make(chan customer)

type customer struct {
	numOfItems int
}

type cashier struct {
	scanSpeed int
}

type till struct {
	numOfItems int
	employee   cashier
	queue      chan customer
}

type manager struct{}

func (t *till) SetUpTill(maxItemsTill bool) {
	if maxItemsTill {
		t.numOfItems = 10
	} else {
		t.numOfItems = 100
	}
	t.employee = cashier{RandomNumberInclusive(1, 10)}
	t.queue = make(chan customer, 6)
}

func (t *till) ProcessCustomer() {
	if len(t.queue) == 0 {
		fmt.Println("queue empty")
	} else {
		currentCustomer := <-t.queue
		fmt.Println(currentCustomer)
	}
}

func (t *till) AddCustomerToQueue(c customer) bool {
	if len(t.queue) == cap(t.queue) {
		fmt.Println("queue full")
		return false
	} else {
		t.queue <- c
		return true
	}
}

func (m *manager) GenerateTills() {
	//generate random number for the num of tills
	max := 8.0
	min := 1.0
	numOfTills := RandomNumberInclusive(min, max)
	tills = make([]till, numOfTills)
	maxItemsTill := 1
	//generate tills and add them to till array and then set them up
	for i := 0; i < numOfTills; i++ {
		//randomly decide if the till has a max number of items
		maxItemsTill = RandomNumberInclusive(1, 100)
		tills[i] = till{}
		if maxItemsTill > 20 {
			tills[i].SetUpTill(false)
		} else {
			tills[i].SetUpTill(true)
		}
	}
}

func RandomNumberInclusive(min, max float64) int {
	return int(min + rand.Float64()*(max-min))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	manager := manager{}
	manager.GenerateTills()
	fmt.Println(tills)
	//test adding customers to queues until they're full
	for i := 0; i < 30; i++ {
		customer := customer{RandomNumberInclusive(1, 100)}
		for j := 0; j < len(tills); j++ {
			if tills[j].AddCustomerToQueue(customer) {
				break
			}
		}
	}
	for j := 0; j < len(tills); j++ {
		fmt.Println(len(tills[j].queue))
	}
}

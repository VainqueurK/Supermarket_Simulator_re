package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

var tills []till
var customers = make([]customer, 100)
var lastCustomerGenerated = time.Now()
var clock = time.Now()
var totalCustomers = 0
var currentNumOfCharacters = 0

type automatic struct {
	running        bool
	generationRate float64
}

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

func (m *manager) GenerateTills() {
	//generate random number for the num of tills
	max := 8.0
	min := 1.0
	numOfTills := randomNumberInclusive(min, max)
	tills = make([]till, numOfTills)

	maxItemsTill := 1
	//generate tills and add them to till array and then set them up
	for i := 0; i < numOfTills; i++ {
		tills[i] = till{}
		//randomly decide if the till has a max number of items
		maxItemsTill = randomNumberInclusive(1, 100)
		if maxItemsTill > 20 {
			tills[i].SetUpTill(false)
		} else {
			tills[i].SetUpTill(true)
		}
	}
}

func (t *till) SetUpTill(maxItemsTill bool) {
	//sets the max num of items. default is 100, but if it's a 'Max item till' it'll be changed to ten
	if maxItemsTill {
		t.numOfItems = 10
	} else {
		t.numOfItems = 100
	}
	//adds cashier with a randomly generated speed to the till
	t.employee = cashier{randomNumberInclusive(1, 10)}
	//the tills queue
	t.queue = make(chan customer, 6)
}

func (t *till) ProcessCustomer() {
	//checks if queue is empty
	if len(t.queue) == 0 {
		fmt.Println("queue empty")
	} else {
		//removes customer from queue
		currentCustomer := <-t.queue
		fmt.Println(currentCustomer)
		//call a method for the cashier to start scanning items
	}
}

func (t *till) AddCustomerToQueue(c customer) bool {
	//checks if queue is full
	if len(t.queue) == cap(t.queue) {
		fmt.Println("queue full")
		return false
	} else {
		//adds customer to queue
		t.queue <- c
		return true
	}
}

func (a *automatic) GenerateCustomers() {
	for a.running {
		time.Sleep(1 * time.Millisecond)
		if time.Now().Sub(lastCustomerGenerated) > (time.Millisecond * 100) {
			customer := customer{randomNumberInclusive(1, 100)}
			customers[currentNumOfCharacters] = customer
			currentNumOfCharacters++
			totalCustomers++
			lastCustomerGenerated = time.Now()
		}
	}
	fmt.Println("wtf why am i here")
}

func (a *automatic) LookForSpaceInQueue() {
	customer := customers[0]
	for i := 0; i < len(tills); i++ {
		if customer.numOfItems <= tills[i].numOfItems {
			//if customer is added break out of the loop
			if tills[i].AddCustomerToQueue(customer) {
				break
			}
		}
	}
}

func (a *automatic) RunSimulator() {
	a.running = true
	manager := manager{}
	manager.GenerateTills()
	go a.GenerateCustomers()
	go a.LookForSpaceInQueue()
	time.Sleep(5 * time.Second)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	//create manager and generate tills
	automatic := automatic{}
	automatic.RunSimulator()

	fmt.Println(tills)
	fmt.Println(fmt.Println(runtime.NumGoroutine()))

	automatic.running = false
	fmt.Println(len(customers))
}

func randomNumberInclusive(min, max float64) int {
	return int(min + rand.Float64()*(max-min))
}

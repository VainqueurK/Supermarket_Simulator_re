package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

var tills []till

//we're gonna have to put this array behind a mutex lock at some point
var customers []customer
var lastCustomerGenerated = time.Now()

//we'll use this in the future
var clock = time.Now()
var totalCustomers = 0
var currentNumOfCustomers = 0

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

func (c *cashier) ScanItems(customer customer) {
	//scanTime := customer.numOfItems * c.scanSpeed - this isn't working atm
	time.Sleep(1000 * time.Millisecond)
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

func (t *till) ProcessCustomer() {
	//checks if queue is empty
	if len(t.queue) == 0 {
		fmt.Println("queue empty")
	} else {
		//removes customer from queue
		currentCustomer := <-t.queue
		fmt.Println(currentCustomer)
		//call a method for the cashier to start scanning items
		t.employee.ScanItems(currentCustomer)
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

func (a *automatic) LookForSpaceInQueue() {
	//wait a few seconds before sorting customers into queues
	time.Sleep(time.Second * 5)
	for a.running {
		customer := customers[0]
		for i := 0; i < len(tills); i++ {
			//if customer num of items is under the max num of items allowed at the till
			if customer.numOfItems <= tills[i].numOfItems {
				//if customer is added break out of the loop
				if tills[i].AddCustomerToQueue(customer) {
					customers = customers[1:]
					currentNumOfCustomers--
					break
				}
			}
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (a *automatic) GenerateCustomers() {
	for a.running {
		time.Sleep(1 * time.Millisecond)
		//if a certain amount of time has passed since the last customer was generated generate a new customer
		if time.Now().Sub(lastCustomerGenerated) > (time.Millisecond * 100) {
			//generate customer
			customer := customer{randomNumberInclusive(1, 100)}
			//add to customer array
			customers = append(customers, customer)
			currentNumOfCustomers++
			totalCustomers++
			lastCustomerGenerated = time.Now()
		}
	}
}

func (a *automatic) RunSimulator() {
	a.running = true
	//create manager agent and generate tills
	manager := manager{}
	manager.GenerateTills()

	//create two goroutines that will continuously generate customers and try to add them to a queue
	go a.GenerateCustomers()
	go a.LookForSpaceInQueue()

	//commented lines don't work yet
	//for i := 0; i < len(tills); i++ {
	//	go tills[i].ProcessCustomer()
	//}
	time.Sleep(5 * time.Second)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	//run simulator
	automatic := automatic{}
	automatic.RunSimulator()

	fmt.Println(tills)
	fmt.Println(fmt.Println(runtime.NumGoroutine()))
	//stop automatic processes
	automatic.running = false
	fmt.Println(len(customers))
}

func randomNumberInclusive(min, max float64) int {
	return int(min + rand.Float64()*(max-min))
}

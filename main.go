/********************************
*		 GROUP Members		*
	Carla Warde - 17204542
	Vainqueur Kayombo - 17199387 
	Vincent Kiely - 17236282
	Áine Reynolds - 17231515
*********************************/
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

/********************************
*		    CONSTANTS			*
*********************************/
const normalQueueMaxNumOfItems = 200
const fastQueueMaxNumOfItems = 20
const maxNumOfTills = 8
const minNumOfTills = 1
const minCashierSpeed = 1
const maxCashierSpeed = 5
const queueLength = 6

/********************************
*	    GLOBAL VARIABLES		*
*********************************/
var tills []till
var hasFastTill bool

//we're gonna have to put this array behind a mutex lock at some point
var customers []customer
var lastCustomerGenerated = time.Now()

//we'll use this in the future
var clock = time.Now()
var totalCustomers = 0
var currentNumOfCustomers = 0
var aineTest = 0
var customersLostDueToImpatients = 0
var running = true

/********************************
*	        STRUCTS				*
*********************************/

type automatic struct {
	generationRate float64
}

type customer struct {
	numOfItems int
	patient bool
}

type cashier struct {
	scanSpeed int
}

type till struct {
	maxNumOfItems int
	employee      cashier
	queue         chan customer
	name          int
}

type manager struct{}

/********************************
*	        METHODS				*
*********************************/

func (a *automatic) RunSimulator() {
	running = true
	//create manager agent and generate tills
	manager := manager{}
	manager.GenerateTills()
	//determine initial generation rate
	a.generationRate = float64(((maxNumOfTills + 1) - len(tills)) * 20)
	//create two goroutines that will continuously generate customers and try to add them to a queue
	go a.GenerateCustomers()
	go a.LookForSpaceInQueue()

	//commented lines don't work yet
	for i := 0; i < len(tills); i++ {
		go tills[i].SendCustomerToCashier()
	}
	//runtime
	time.Sleep(60 * time.Second)
}

func (m *manager) GenerateTills() {
	//generate random number for the num of tills
	numOfTills := randomNumberInclusive(minNumOfTills, maxNumOfTills)
	tills = make([]till, numOfTills)

	index := 0
	//guarantee fast till is generated if numOfTills > 1, and that if numOfTills == 1 it's a regular till
	if numOfTills > 1 {
		//guaranteed fast till
		tills[0] = till{name: 1}
		tills[0].SetUpTill(true)
		index++
		//guaranteed regular till
		tills[1] = till{name: 2}
		tills[1].SetUpTill(false)
		index++
		hasFastTill = true
	} else {
		//guaranteed regular till
		tills[0] = till{name: 2}
		tills[0].SetUpTill(false)
		index++
		hasFastTill = false
	}

	maxItemsTill := 1
	//generate the rest of the tills randomly
	for i := index; i < numOfTills; i++ {
		tills[i] = till{name: (i + 1)}
		//randomly decide if the till has a max number of items
		maxItemsTill = randomNumberInclusive(1, 100)
		if maxItemsTill > 10 {
			tills[i].SetUpTill(false)
		} else {
			tills[i].SetUpTill(true)
		}
	}
}

func (t *till) SetUpTill(maxItemsTill bool) {
	//sets the max num of items. default is 200, but if it's a 'Max item till' it'll be changed to 20
	if maxItemsTill {
		t.maxNumOfItems = fastQueueMaxNumOfItems
	} else {
		t.maxNumOfItems = normalQueueMaxNumOfItems
	}
	//adds cashier with a randomly generated speed to the till
	t.employee = cashier{randomNumberInclusive(minCashierSpeed, maxCashierSpeed)}
	//the tills queue
	t.queue = make(chan customer, queueLength)
}

func (a *automatic) GenerateCustomers() {
	for running {
		time.Sleep(10 * time.Millisecond)
		//if a certain amount of time has passed since the last customer was generated generate a new customer
		if time.Now().Sub(lastCustomerGenerated) > (time.Millisecond * time.Duration(a.generationRate)) {
			//generate customer
			var patient bool
			patientInt := randomNumberInclusive(1,3)
			if patientInt == 1 {
				patient = true
			} else {
				patient = false
			}
			customer := customer{randomNumberInclusive(1, 200), patient}
			//fmt.Printf("Customer patient attribute: %t\n", customer.patient)
			//add to customer array
			customers = append(customers, customer)
			currentNumOfCustomers++
			aineTest++
			totalCustomers++
			lastCustomerGenerated = time.Now()
		}
	}
}

func (a *automatic) LookForSpaceInQueue() {
	var index int
	for running {
		time.Sleep(20 * time.Millisecond)
		//check if customers are waiting
		if len(customers) > 0 {
			customer := customers[0]
			//patient := customer.patient
			//checks if customer can use fast queue
			if customer.numOfItems <= fastQueueMaxNumOfItems && hasFastTill {
				//find fast queue index
				index = shortestFastQueue()
			} else {
				//find shortest normal queue index
				index = shortestAvailableQueue(customer.numOfItems)
			}

			//if no queue is found index == -1
			if index == -1 {
				//fmt.Println("no available queue")
				//logic for if there's no queue available for the customer
			} else {
				//if customer is added remove customer from array
				if tills[index].AddCustomerToQueue(customer) {
					customers = customers[1:]
					currentNumOfCustomers--
				}
			}
		}
	}
}

func shortestAvailableQueue(numOfItems int) int {
	min := queueLength
	var tillIndex = -1
	//loop through array and find till with the shortest queue that the customer can go to
	for i := 0; i < len(tills); i++ {
		if len(tills[i].queue) < min && numOfItems <= tills[i].maxNumOfItems {
			min = len(tills[i].queue)
			tillIndex = i
		}
	}
	return tillIndex
}

func shortestFastQueue() int {
	min := queueLength
	index := -1
	for i := 0; i < len(tills); i++ {
		if len(tills[i].queue) < min && tills[i].maxNumOfItems == fastQueueMaxNumOfItems {
			min = len(tills[i].queue)
			index = i
		}
	}
	return index
}

func (t *till) SendCustomerToCashier() {
	for running {
		//a wait time so the loop doesn't run too fast
		time.Sleep(30 * time.Millisecond)
		//checks if queue is empty
		if len(t.queue) == 0 {
			//fmt.Println("queue empty")
		} else {
			//removes customer from queue
			currentCustomer := <-t.queue
			fmt.Printf("Scanning %d items in Till %d\n", currentCustomer.numOfItems, t.name)
			//call a method for the cashier to start scanning items
			t.employee.ScanItems(currentCustomer)
		}
	}
}

func (t *till) AddCustomerToQueue(c customer) bool {
	//checks if queue is full
	if len(t.queue) == cap(t.queue) {
		//fmt.Println("queue full")
		return false
	} else {
		//add logic for impatient customer
		tillNum := len(tills)
		//impatient customer will leave if there is more than 2 people in each queue
		numOfCustomerForImpatientToLeave := tillNum * 2
		if aineTest > numOfCustomerForImpatientToLeave && c.patient == false {
			fmt.Println("Customer is impatient and leaving")
			customersLostDueToImpatients++
			aineTest--
		}else{
			//adds customer to queue
			t.queue <- c
		}
		return true
	}
}

func (c *cashier) ScanItems(customer customer) {
	scanTime := customer.numOfItems * c.scanSpeed
	time.Sleep(time.Duration(scanTime) * time.Millisecond)
	aineTest--
}

/********************************
*	        FUNCTIONS			*
*********************************/

func main() {
	rand.Seed(time.Now().UnixNano())
	//run simulator
	automatic := automatic{}
	automatic.RunSimulator()
	fmt.Println(tills)
	fmt.Println(fmt.Println(runtime.NumGoroutine()))
	//stop automatic processes
	running = false
	fmt.Printf("Current customers: %d\n", currentNumOfCustomers)
	fmt.Printf("Total number of customers: %d\n", totalCustomers)
	fmt.Printf("Total number of impatient customers lost: %d\n", customersLostDueToImpatients)
}

func randomNumberInclusive(min, max float64) int {
	num := min + rand.Float64()*(max-min)
	//fmt.Println(num, int(num))
	return int(num)
}

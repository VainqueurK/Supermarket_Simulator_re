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
const (
	normalQueueMaxNumOfItems = 200
	fastQueueMaxNumOfItems   = 20
	maxNumOfTills            = 8
	minNumOfTills            = 1
	minCashierSpeed          = 1
	maxCashierSpeed          = 5
	queueLength              = 6
)

//Days of the week variables
const (
	MONDAY    = 1.25
	TUESDAY   = 1.5
	WEDNESDAY = 2
	THURSDAY  = 1
	FRIDAY    = 0.8
	SATURDAY  = 0.7
	SUNDAY    = 0.5
)

/********************************
*	    GLOBAL VARIABLES		*
*********************************/
var tills []till
var hasFastTill bool
var daysOfTheWeek = [...]string{"a", "b", "c", "d", "e", "f", "g"}

//we're gonna have to put this array behind a mutex lock at some point
var customers []customer
var lastCustomerGenerated = time.Now()

//we'll use this in the future
var clock = time.Now()
var totalCustomers = 0
var currentNumOfCustomers = 0
var running = true
var timeRunning = 0
var day float64 = 0

/********************************
*	        STRUCTS				*
*********************************/

type automatic struct {
	generationRate float64
}

type customer struct {
	numOfItems int
}

type cashier struct {
	scanSpeed int
}

type till struct {
	maxNumOfItems int
	employee      cashier
	queue         chan customer
	name          int
	scannedItems  int
	tillUsage     int
}

type manager struct{}

/********************************
*	        METHODS				*
*********************************/

func (a *automatic) RunSimulator() {
	running = true
	//Get user input for day and runtime
	getInputs()
	//create manager agent and generate tills
	manager := manager{}
	manager.GenerateTills()
	//determine initial generation rate
	a.generationRate = float64((((maxNumOfTills + 1) - len(tills)) * 20))
	a.generationRate = a.generationRate * day
	//create two goroutines that will continuously generate customers and try to add them to a queue
	go a.GenerateCustomers()
	go a.LookForSpaceInQueue()

	//commented lines don't work yet
	for i := 0; i < len(tills); i++ {
		go tills[i].SendCustomerToCashier()
	}
	//runtime
	time.Sleep(time.Duration(timeRunning) * time.Second)
}

func getInputs() {
	var input string
	valid := false
	for !valid {
		//Get inputs from the command line to decide date
		fmt.Println("Enter the Day of the Week you wish to simulate: ")
		fmt.Println("a)Monday b)Tuesday c)Wednesday d)Thursday e)Friday f)Saturday g)Sunday")
		fmt.Scanln(&input)
		switch input {
		case daysOfTheWeek[0]:
			fmt.Println("The chosen day is Monday")
			valid = true
			day = MONDAY
		case daysOfTheWeek[1]:
			fmt.Println("The chosen day is Tuesday")
			valid = true
			day = TUESDAY
		case daysOfTheWeek[2]:
			fmt.Println("The chosen day is Wednesday")
			valid = true
			day = WEDNESDAY
		case daysOfTheWeek[3]:
			fmt.Println("The chosen day is Thursday")
			valid = true
			day = THURSDAY
		case daysOfTheWeek[4]:
			fmt.Println("The chosen day is Friday")
			valid = true
			day = FRIDAY
		case daysOfTheWeek[5]:
			fmt.Println("The chosen day is Saturday")
			valid = true
			day = SATURDAY
		case daysOfTheWeek[6]:
			fmt.Println("The chosen day is Sunday")
			valid = true
			day = SUNDAY
		default:
			fmt.Println("Error: Invalid Input Detected")
			fmt.Println("Example Usage: Enter a for Monday")
		}
	}
	fmt.Println("Enter the number of seconds you want the program to run for: ")
	fmt.Scanln(&timeRunning)
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
		tills[0] = till{name: 1}
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
			customer := customer{randomNumberInclusive(1, 200)}
			//add to customer array
			customers = append(customers, customer)
			currentNumOfCustomers++
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
			t.tillUsage++
			t.scannedItems += currentCustomer.numOfItems
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

		//adds customer to queue
		t.queue <- c
		return true
	}
}

func (c *cashier) ScanItems(customer customer) {
	scanTime := customer.numOfItems * c.scanSpeed
	time.Sleep(time.Duration(scanTime) * time.Millisecond)
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
	for i := 0; i < len(tills); i++ {
		fmt.Printf("Number of items scanned by till %d: %d\n", tills[i].name, tills[i].scannedItems)
		fmt.Printf("Number of customers processed by till %d: %d\n", tills[i].name, tills[i].tillUsage)
	}
}

func randomNumberInclusive(min, max float64) int {
	num := min + rand.Float64()*(max-min)
	//fmt.Println(num, int(num))
	return int(num)
}

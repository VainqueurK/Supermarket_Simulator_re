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
	//"strconv"
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
	minCashierSpeed          = 1.0
	maxCashierSpeed          = 2.0
	queueLength              = 6
	timeRunning 			 = 30
	millisecPerRealHour 	 = float64((float64(timeRunning)/ 12) * 1000)
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

var customers []customer
var lastCustomerGenerated = time.Now()
var lastTillChanged = time.Now()

var clock = time.Now()
var totalCustomers = 0
var customerCount = 0
var currentNumOfCustomers = 0
var numOfCutomersInShop = 0
var numOfOpenTills = 0
var customersLostDueToImpatients = 0
var running = true
var day float64 = 0
var avgWaitTime float64 = 0

/********************************
*	        STRUCTS				*
*********************************/

type automatic struct {
	generationRate float64
}

type customer struct {
	numOfItems int
	patient bool
	stime time.Time
	waitTime float64
}

type cashier struct {
	scanSpeed float64
}

type till struct {
	maxNumOfItems int
	employee      cashier
	queue         chan customer
	name          int
	open bool
	scannedItems  int
	tillUsage     int
	lastUsed	time.Time
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
	twoMinInMilli := (millisecPerRealHour/60) * 2
	a.generationRate = float64(twoMinInMilli)
	a.generationRate = a.generationRate * day 
	//a.generationRate = a.generationRate * thirtySecInMilli
	//create two goroutines that will continuously generate customers and try to add them to a queue
	go a.GenerateCustomers()
	go a.LookForSpaceInQueue()

	//commented lines don't work yet
	for i := 0; i < len(tills); i++ {
		go tills[i].SendCustomerToCashier()
	}

	//Create a go rountine that consistently checks if tills should be opened or closed depending on the number of people in the supermarket
	time.Sleep(time.Duration(20) * time.Millisecond)
	go a.OpenTillIfBusy()
	go a.CloseTills()
	//runtime
	time.Sleep(time.Duration(timeRunning) * time.Second)
}

func getInputs() {
	var input string
	valid := false

	//Runs for loop until there is a valid input for the day
	for !valid {
		//Get inputs from the command line to decide date
		fmt.Println("Enter the Day of the Week you wish to simulate: ")
		fmt.Println("a)Monday b)Tuesday c)Wednesday d)Thursday e)Friday f)Saturday g)Sunday")
		fmt.Scanln(&input)
		//Uses a switch statement to go through the different cases to set the day the simulator will simulate
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
}

func (m *manager) GenerateTills() {
	//generate random number for the num of tills
	numOfTills := int(randomNumberInclusive(2, maxNumOfTills))
	tills = make([]till, maxNumOfTills)

	index := 0
	//guaranteed fast till
	tills[0] = till{name: 1}
	tills[0].SetUpTill(true)
	index++
	//guaranteed regular till
	tills[1] = till{name: 2}
	tills[1].SetUpTill(false)
	index++
	hasFastTill = true

	//maxItemsTill := 1
	//generate the rest of the tills as regular tills
	for i := index; i < maxNumOfTills; i++ {
		tills[i] = till{name: (i + 1)}
		tills[i].SetUpTill(false)
	}

	for i := 0; i < numOfTills; i++ {
		tills[i].open = true
		numOfOpenTills++
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
	t.open = false
}

func (a *automatic) GenerateCustomers() {
	for running {
		//if a certain amount of time has passed since the last customer was generated generate a new customer
		if time.Now().Sub(lastCustomerGenerated) > (time.Millisecond * time.Duration(a.generationRate)) {
			//fmt.Printf("Time now - last customer generated = %v\n",time.Now().Sub(lastCustomerGenerated))
			//generate customer
			var patient bool
			patientInt := randomNumberInclusive(1,3)
			if patientInt == 1 {
				patient = true
			} else {
				patient = false
			}

			customer := customer{int(randomNumberInclusive(1, 200)), patient, time.Now() , 0.0}
			//fmt.Printf("Customer patient attribute: %t\n", customer.patient)
			//add to customer array
			customers = append(customers, customer)
			currentNumOfCustomers++
			numOfCutomersInShop++
			totalCustomers++
			lastCustomerGenerated = time.Now()
		}
	}
}

func (a *automatic) LookForSpaceInQueue() {
	var index int
	for running {
		//Check for space every minute
		fiveSecInMilli := ((millisecPerRealHour/60)/60) * 5
		time.Sleep(time.Duration(fiveSecInMilli) * time.Millisecond)
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
				if len(tills[index].queue) > 3   && customer.patient == false {
					customersLostDueToImpatients++
					numOfCutomersInShop--
					customers = customers[1:]
					currentNumOfCustomers--
					continue
				}else{
					tills[index].AddCustomerToQueue(customer) 
					customers = customers[1:]
					currentNumOfCustomers--
				}
			}
		}
	}
}

func (a *automatic) OpenTillIfBusy(){
	//check tills every 20 minutes  
	twentyMinInMilli := (millisecPerRealHour/60) * 20
	for running {
		if time.Now().Sub(lastTillChanged) >  (time.Millisecond * time.Duration(twentyMinInMilli)){
			lenOfQueues := 0
			numTills := 0
			for i:=0; i < numOfOpenTills; i++{
				if tills[i].maxNumOfItems == 200{
					lenOfQueues += len(tills[i].queue)
					numTills++
				}
			}
			lenOfQueues = lenOfQueues/numTills
			if numOfOpenTills < 8 && lenOfQueues > 3 {
				fmt.Printf("Time difference between now and last Till changed = %v\n", time.Now().Sub(lastTillChanged))
				tills[numOfOpenTills].open = true
				tills[numOfOpenTills].lastUsed = time.Now()
				tills[numOfOpenTills].maxNumOfItems = 200
				numOfOpenTills++
				fmt.Printf("Opening another Till num %d\n", tills[numOfOpenTills-1].name)
				lastTillChanged = time.Now()
			}
		}
	}
}

func (a *automatic) CloseTills(){
	for running {	
		twentyMinInMilli := (millisecPerRealHour/60) * 20
		fortyMinInMilli := (millisecPerRealHour/60) * 40
		if time.Now().Sub(lastTillChanged) >  (time.Millisecond * time.Duration(fortyMinInMilli)) {
			if numOfOpenTills > 2 &&  time.Now().Sub(tills[numOfOpenTills-1].lastUsed) > (time.Millisecond * time.Duration(twentyMinInMilli)){
				tills[numOfOpenTills-1].open = false
				numOfOpenTills--
				fmt.Printf("Closing a Till num %d\n", tills[numOfOpenTills].name)
				lastTillChanged = time.Now()
			}
		}
	}
}

func shortestAvailableQueue(numOfItems int) int {
	min := queueLength
	var tillIndex = -1
	//loop through array and find till with the shortest queue that the customer can go to
	for i := 0; i < len(tills); i++ {
		if len(tills[i].queue) < min && numOfItems <= tills[i].maxNumOfItems && tills[i].open == true{
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
		if len(tills[i].queue) < min && tills[i].maxNumOfItems == fastQueueMaxNumOfItems && tills[i].open == true{
			min = len(tills[i].queue)
			index = i
		}
	}
	return index
}

func (t *till) SendCustomerToCashier() {
	for running {
		//a wait time so the loop doesn't run too fast
		tensec := (((millisecPerRealHour)/60)/60) * 10
		time.Sleep(time.Duration(tensec) * time.Millisecond)
		//checks if queue is empty
		if len(t.queue) == 0 {
			//fmt.Println("queue empty")
		} else {
			
			//removes customer from queue
			currentCustomer := <-t.queue

			customerCount++
			endTime := time.Now()
			waitT := float64(endTime.Sub(currentCustomer.stime).Milliseconds())
			
			// Get the actual customer wait time in relation to a day.
			custWaitT := waitT * millisecPerRealHour
			//fmt.Printf("Time running = %d waitT = %f milliPerSec = %f\n", timeRunning,waitT, milliPerRealSec)
			//Get each customer wait time in minutes and assign it to the customer
			currentCustomer.waitTime = (custWaitT / 1000)/60
			//Add up all the wait times to use to get the average
			avgWaitTime = float64(avgWaitTime) + currentCustomer.waitTime
			fmt.Printf("Customer num %d wait time = %f minutes in till: %d\n", customerCount, currentCustomer.waitTime, t.name)
			//fmt.Printf("Scanning %d items in Till %d\n", currentCustomer.numOfItems, t.name)
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

		//adds customer to queue
		t.queue <- c
		t.lastUsed = time.Now()
		return true
	}
}

func (c *cashier) ScanItems(customer customer) {
	threeSec := (((millisecPerRealHour)/60)/60) * 3
	scanTime := float64(customer.numOfItems) * c.scanSpeed * threeSec
	time.Sleep(time.Duration(scanTime) * time.Millisecond)
	numOfCutomersInShop--
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
	fmt.Printf("Total number of tills open: %d\n", numOfOpenTills)
	fmt.Printf("Total number of impatient customers lost: %d\n", customersLostDueToImpatients)
	fmt.Printf("Average customer wait time: %d minutes\n", (int(avgWaitTime)/customerCount))
	fmt.Printf("Running time = %d Milliseconds in an hour = %f\n", timeRunning, millisecPerRealHour)
	for i := 0; i < len(tills); i++ {
		fmt.Printf("Number of items scanned by till %d: %d\n", tills[i].name, tills[i].scannedItems)
		fmt.Printf("Number of customers processed by till %d: %d\n", tills[i].name, tills[i].tillUsage)
	}
}

func randomNumberInclusive(min, max float64) float64 {
	num := min + rand.Float64()*(max-min)
	//fmt.Println(num, int(num))
	return float64(num)
}

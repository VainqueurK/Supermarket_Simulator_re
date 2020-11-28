/********************************
*		 GROUP Members		*
	Group Name = Clean up on aisle 3
	Carla Warde - 17204542
	Vainqueur Kayombo - 17199387 
	Vincent Kiely - 17236282
	Áine Reynolds - 17231515
*********************************/
package main

import (
	"fmt"
	"math/rand"
	"time"
	"os"
)

/********************************
*		    CONSTANTS			*
*********************************/
const (
	minNumOfItems			 = 1
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
//Weather variables
const (
	DRIZZLE 	= 1.1
	HEAVYRAIN	= 1.5
	SUNNY 		= 0.5
	CLOUDY 		= 0.7
	SNOWY 		= 1.2
)
/********************************
*	    GLOBAL VARIABLES		*
*********************************/
var tills []till
var hasFastTill bool
var verbose = false
var daysOfTheWeek = [...]string{"a", "b", "c", "d", "e", "f", "g"}
var weather = [...]string{"a", "b", "c", "d", "e"}

var customers []customer
var lastCustomerGenerated = time.Now()
var lastTillChanged = time.Now()

var clock = time.Now()
var totalCustomers = 0
var customerCount = 0
var currentNumOfCustomers = 0
var numOfCutomersInShop = 0
var numOfOpenTills = 0
var lostCustomers = 0
var customersLostDueToImpatience = 0
var running = true
var day float64 = 0
var avgWaitTime float64 = 0

var weatherDelay float64 // V
var dayDelay float64     //run method to set this when simulator is run
var delay float64        //dayDelay * weatherDelay

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
	delay = dayDelay * weatherDelay
	twoMinInMilli := (millisecPerRealHour/60) * 2 
	a.generationRate = float64(twoMinInMilli)
	//fmt.Printf("Gen rate before: %f\n", a.generationRate)
	a.generationRate = a.generationRate * delay
	//fmt.Printf("Gen rate after: %f\n", a.generationRate)

	//create two goroutines that will continuously generate customers and try to add them to a queue
	go a.GenerateCustomers()
	go a.LookForSpaceInQueue()

	//Create go routines for each till
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
		dayDelay = setDay(input)
		if dayDelay != -1.0 {
			valid = true
		}
	}

	valid = false
	for !valid {
		//Get inputs from the command line to decide date
		fmt.Println("Enter the Day of the Week you wish to simulate: ")
		fmt.Println("a)Cloudy b)Sunny c)Drizzly d)Heavy Rain e)Snowy")
		fmt.Scanln(&input)
		//Uses a switch statement to go through the different cases to set the day the simulator will simulate
		weatherDelay = setWeather(input)
		if weatherDelay != -1.0 {
			valid = true
		}
	}
}

func setDay(input string) float64 {
	var delay = -1.0

	switch input {
		case daysOfTheWeek[0]:
			fmt.Println("The chosen day is Monday")
			delay = MONDAY
		case daysOfTheWeek[1]:
			fmt.Println("The chosen day is Tuesday")
			delay = TUESDAY
		case daysOfTheWeek[2]:
			fmt.Println("The chosen day is Wednesday")
			delay = WEDNESDAY
		case daysOfTheWeek[3]:
			fmt.Println("The chosen day is Thursday")
			delay = THURSDAY
		case daysOfTheWeek[4]:
			fmt.Println("The chosen day is Friday")
			delay = FRIDAY
		case daysOfTheWeek[5]:
			fmt.Println("The chosen day is Saturday")
			delay = SATURDAY
		case daysOfTheWeek[6]:
			fmt.Println("The chosen day is Sunday")
			delay = SUNDAY
		default:
			fmt.Println("Error: Invalid Input Detected")
			fmt.Println("Example Usage: Enter 'a' for Monday")
	}
	return delay
}

func setWeather(input string) float64 {
	var wDelay = -1.0
	switch input {
		case weather[0]:
			fmt.Println("The chosen weather is 'Cloudy'")
			wDelay = CLOUDY
		case weather[1]:
			fmt.Println("The chosen weather is 'Sunny'")
			wDelay = SUNNY
		case weather[2]:
			fmt.Println("The chosen weather is 'Drizzly'")
			wDelay = DRIZZLE
		case weather[3]:
			fmt.Println("The chosen weather is 'Heavy Rain'")
			wDelay = HEAVYRAIN
		case weather[4]:
			fmt.Println("The chosen weather is 'Snowy'")
			wDelay = SNOWY
		default:
			fmt.Println("Error: Invalid Input Detected")
			fmt.Println("Example Usage: Enter 'a' for Cloudy")
		}
	return wDelay
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

	//generate the rest of the tills as regular tills
	for i := index; i < maxNumOfTills; i++ {
		tills[i] = till{name: (i + 1)}
		tills[i].SetUpTill(false)
	}

	//Open random number of tills
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
			//Generate odds of customer being impatient
			var patient bool
			patientInt := randomNumberInclusive(1,3)
			if patientInt > 1 {
				patient = true
			} else {
				patient = false
			}

			//Generate customer with random number of items
			customer := customer{int(randomNumberInclusive(minNumOfItems, normalQueueMaxNumOfItems)), patient, time.Now() , 0.0}
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
		//Check for space every 5 seconds
		fiveSecInMilli := ((millisecPerRealHour/60)/60) * 5
		time.Sleep(time.Duration(fiveSecInMilli) * time.Millisecond)
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

			//if no queue has less than 6 we lose the customer
			if index == -1 {
				lostCustomers++
				numOfCutomersInShop--
				customers = customers[1:]
				currentNumOfCustomers--
			} else {
				//if customer is impatient and the length of the shortest queue is 4 they leave
				if len(tills[index].queue) > 3   && customer.patient == false {
					customersLostDueToImpatience++
					numOfCutomersInShop--
					customers = customers[1:]
					currentNumOfCustomers--
					continue
				}else{
					//Send customer to queue
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
			//Get average length of the queues
			for i:=0; i < numOfOpenTills; i++{
				if tills[i].maxNumOfItems == 200{
					lenOfQueues += len(tills[i].queue)
					numTills++
				}
			}
			lenOfQueues = lenOfQueues/numTills
			//If average length of queues is above three open till
			if numOfOpenTills < 8 && lenOfQueues > 3 {
				//fmt.Printf("Time difference between now and last Till changed = %v\n", time.Now().Sub(lastTillChanged))
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
			//Close till if it wasn't used in the last 20 min
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
	//loop through array and find till with the shortest fast queue that the customer can go to
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
			if verbose{
				fmt.Printf("Customer num %d wait time = %f minutes in till: %d\n", customerCount, currentCustomer.waitTime, t.name)
			}			
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
	//1 item takes 2 seconds to scan multiplied by cashier scan speed
	twoSec := (((millisecPerRealHour)/60)/60) * 2
	scanTime := float64(customer.numOfItems) * c.scanSpeed * twoSec
	time.Sleep(time.Duration(scanTime) * time.Millisecond)
	numOfCutomersInShop--
}

func (t *till) String() string{
	return fmt.Sprintf("Till %d stats:\n\tTill open: %t\n\tMax num of items: %d\n\t" +
						"Cashier Scan speed: %.2f\n\tNumber of Items scanned: %d\n\t" +
						"Number of customers processed: %d\n", 
						t.name, t.open, t.maxNumOfItems, t.employee.scanSpeed, t.scannedItems, t.tillUsage)
}

/********************************
*	        FUNCTIONS			*
*********************************/

func main() {
	//Check if user wants verbose to print customer wait times
	args := os.Args[1:]
	if args != nil {
		if args[0] == "-v"{
			fmt.Println("Verbose mode selected")
			verbose = true
		} 
	}
	
	rand.Seed(time.Now().UnixNano())
	//run simulator
	automatic := automatic{}
	automatic.RunSimulator()
	//stop automatic processes
	running = false
	time.Sleep(20 * time.Millisecond)
	fmt.Printf("Current customers: %d\n", currentNumOfCustomers)
	fmt.Printf("Total number of customers: %d\n", totalCustomers)
	fmt.Printf("Total number of tills open: %d\n", numOfOpenTills)
	fmt.Printf("Total number of patient customers lost: %d\n", lostCustomers)
	fmt.Printf("Total number of impatient customers lost: %d\n", customersLostDueToImpatience)
	fmt.Printf("Average customer wait time: %d minutes\n", (int(avgWaitTime)/customerCount))
	for i := 0; i < len(tills); i++ {
		fmt.Println(tills[i].String())
	}
}

func randomNumberInclusive(min, max float64) float64 {
	num := min + rand.Float64()*(max-min)
	//fmt.Println(num, int(num))
	return float64(num)
}

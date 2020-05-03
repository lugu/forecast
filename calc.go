package main

import (
	"flag"
	"fmt"
	"time"
)

type Shipment struct {
	arrival  int
	quantity int
}

type Parameters struct {
	cash          float64
	batch         int
	unitCost      float64
	unitBenefit   float64
	weeklyRate    float64
	shipmentDelay int
	duration      int
}

type Simulation struct {
	Date  []time.Time
	Stock []int
	Cash  []float64
	Param Parameters
}

func NewSimulation(param Parameters) Simulation {
	stock := 0.0
	cash := param.cash
	shipments := []Shipment{}

	sim := Simulation{
		Date:  make([]time.Time, param.duration),
		Stock: make([]int, param.duration),
		Cash:  make([]float64, param.duration),
	}

	for day := 0; day < param.duration; day++ {
		date := time.Now().AddDate(0, 0, day)

		sim.Date[day] = date
		sim.Stock[day] = int(stock)
		sim.Cash[day] = cash

		sellRate := param.weeklyRate / 7.0
		batchCost := float64(param.batch) * param.unitCost

		// if we can buy some more stock
		for cash >= batchCost {

			// optimal stock: two time the shipment delay
			runway := sellRate * float64(param.shipmentDelay) * 2

			// compute the pending amount before buying
			// some more.
			pending := 0
			for _, shipment := range shipments {
				pending += shipment.quantity
			}

			// buy stock if we need some more
			if stock+float64(pending) < runway {
				shipments = append(shipments, Shipment{
					arrival:  day + param.shipmentDelay,
					quantity: param.batch,
				})
				cash -= batchCost
			} else {
				// otherwise, don't buy.
				break
			}
		}
		for {
			if len(shipments) == 0 {
				break
			}
			first := shipments[0]
			if day == first.arrival {
				shipments = shipments[1:]
				stock += float64(first.quantity)
			} else {
				break
			}
		}
		if stock > sellRate {
			stock -= sellRate
			unitGain := param.unitCost + param.unitBenefit
			cash += sellRate * unitGain
		} else {
			unitGain := param.unitCost + param.unitBenefit
			cash += stock * unitGain
			stock = 0
		}
	}
	return sim
}

func (s Simulation) Print() {
	fmt.Printf("cash\t%.2f\tinitial investment\n", s.Param.cash)
	fmt.Printf("rate\t%.2f\tweekly sales\n", s.Param.weeklyRate)
	fmt.Printf("cost\t%.2f\tprice of each unit\n", s.Param.unitCost)
	fmt.Printf("gain\t%.2f\tmargin for each unit\n", s.Param.unitBenefit)
	fmt.Printf("delay\t%d\tdays to ship\n", s.Param.shipmentDelay)
	fmt.Printf("\n")
	fmt.Printf("day\tcash\tstock\n")
	for i := range s.Date {
		fmt.Printf("%v\t%.2f\t%d\n", s.Date[i].Format("01-02-2006"), s.Cash[i], s.Stock[i])
	}
}

func main() {
	var param Parameters

	param.cash = 1000.0
	param.batch = 20
	param.unitCost = 25.0
	param.unitBenefit = 10.0
	param.weeklyRate = 7.0
	param.shipmentDelay = 14
	param.duration = 365

	flag.Float64Var(&param.cash, "cash", param.cash, "initial investment (USD)")
	flag.Float64Var(&param.weeklyRate, "rate", param.weeklyRate, "average sells per week (quantity)")
	flag.Float64Var(&param.unitCost, "cost", param.unitCost, "cost of each unit (USD)")
	flag.Float64Var(&param.unitBenefit, "gain", param.unitBenefit, "gain for each unit (USD)")
	flag.IntVar(&param.batch, "batch", param.batch, "size of each shipment (quantity)")
	flag.IntVar(&param.shipmentDelay, "delay", param.shipmentDelay, "time to ship a batch (days)")
	flag.IntVar(&param.duration, "days", param.duration, "simulation duration (days)")
	flag.Parse()

	sim := NewSimulation(param)
	sim.Print()
}

package main

import (
	"flag"
	"fmt"
)

type container struct {
	arrival  int
	quantity int
}

func main() {

	cash := 1000.0
	batch := 20
	unitCost := 25.0
	unitBenefit := 10.0
	weeklyRate := 7.0
	shipmentDelay := 14
	duration := 365

	flag.Float64Var(&cash, "cash", cash, "initial investment (USD)")
	flag.Float64Var(&weeklyRate, "rate", weeklyRate, "average sells per week (quantity)")
	flag.Float64Var(&unitCost, "cost", unitCost, "cost of each unit (USD)")
	flag.Float64Var(&unitBenefit, "gain", unitBenefit, "gain for each unit (USD)")
	flag.IntVar(&batch, "batch", batch, "size of each shipment (quantity)")
	flag.IntVar(&shipmentDelay, "delay", shipmentDelay, "time to ship a batch (days)")
	flag.IntVar(&duration, "days", duration, "simulation duration (days)")
	flag.Parse()

	stock := 0.0
	shipments := []container{}

	fmt.Printf("cash\t%.2f\tinitial investment\n", cash)
	fmt.Printf("rate\t%.2f\tweekly sales\n", weeklyRate)
	fmt.Printf("cost\t%.2f\tprice of each unit\n", unitCost)
	fmt.Printf("gain\t%.2f\tmargin for each unit\n", unitBenefit)
	fmt.Printf("delay\t%d\tdays to ship\n", shipmentDelay)
	fmt.Printf("\n")
	fmt.Printf("day\tcash\tstock\n")

	for day := 0; day < duration; day++ {

		sellRate := weeklyRate / 7.0

		batchCost := float64(batch) * unitCost

		// if we can buy some more stock
		for cash >= batchCost {

			// optimal stock: two time the shipment delay
			runway := sellRate * float64(shipmentDelay) * 2

			// compute the pending amount before buying
			// some more.
			pending := 0
			for _, shipment := range shipments {
				pending += shipment.quantity
			}

			// buy stock if we need some more
			if stock+float64(pending) < runway {
				shipments = append(shipments, container{arrival: day + shipmentDelay, quantity: batch})
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
			unitGain := unitCost + unitBenefit
			cash += sellRate * unitGain
		} else {
			unitGain := unitCost + unitBenefit
			cash += stock * unitGain
			stock = 0
		}

		fmt.Printf("%d\t%.2f\t%d\n", day, cash, int(stock))
	}
}

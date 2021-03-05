package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"time"

	ui "github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/style"
	"github.com/wcharczuk/go-chart"
)

type Shipment struct {
	arrival  int
	quantity int
}

type Parameters struct {
	cash               float64
	batch              int
	unitCost           float64
	unitBenefit        float64
	unitMonthlyStorage float64
	weeklySales        float64
	shipmentDelay      int
	duration           int
}

type Simulation struct {
	Date  []time.Time
	Stock []float64
	Cash  []float64
	Param Parameters
}

var (
	sim Simulation
)

func NewSimulation(param Parameters) Simulation {
	stock := 0.0
	cash := param.cash
	shipments := []Shipment{}

	sim := Simulation{
		Date:  make([]time.Time, param.duration),
		Stock: make([]float64, param.duration),
		Cash:  make([]float64, param.duration),
		Param: param,
	}

	for day := 0; day < param.duration; day++ {
		date := time.Now().AddDate(0, 0, day)

		sim.Date[day] = date
		sim.Stock[day] = stock
		sim.Cash[day] = cash

		sellRate := param.weeklySales / 7.0
		batchCost := float64(param.batch) * param.unitCost

		// storage cost
		unitDailyStorage := param.unitMonthlyStorage / 30
		cash -= stock * unitDailyStorage

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
	fmt.Printf("sales\t%.2f\tweekly sales\n", s.Param.weeklySales)
	fmt.Printf("storage\t%.2f\tstorage cost per unit per month\n", s.Param.unitMonthlyStorage)
	fmt.Printf("cost\t%.2f\tprice of each unit\n", s.Param.unitCost)
	fmt.Printf("margin\t%.2f\tmargin for each unit\n", s.Param.unitBenefit)
	fmt.Printf("delay\t%d\tdays to ship\n", s.Param.shipmentDelay)
	fmt.Printf("\n")
	fmt.Printf("day\tcash\tstock\n")
	for i := range s.Date {
		fmt.Printf("%v\t%.2f\t%d\n", s.Date[i].Format("01-02-2006"), s.Cash[i], int(s.Stock[i]))
	}
}

func (s Simulation) Plot() (image.Image, error) {

	graph := chart.Chart{
		Title: "Stock vs Cash",
		YAxis: chart.YAxis{
			Name: "USD",
		},
		YAxisSecondary: chart.YAxis{
			Name: "Stock",
		},
		XAxis: chart.XAxis{
			TickPosition:   chart.TickPositionBetweenTicks,
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Stock available",
				Style: chart.Style{
					StrokeColor: chart.GetDefaultColor(1).WithAlpha(64),
					StrokeWidth: 2.0,
				},
				XValues: s.Date,
				YValues: s.Stock,
				YAxis:   chart.YAxisSecondary,
			},
			chart.TimeSeries{
				Name: "Cash available",
				Style: chart.Style{
					StrokeColor: chart.GetDefaultColor(2).WithAlpha(64),
					StrokeWidth: 2.0,
				},
				XValues: s.Date,
				YValues: s.Cash,
			},
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	collector := &chart.ImageWriter{}
	err := graph.Render(chart.PNG, collector)
	if err != nil {
		return nil, err
	}
	return collector.Image()
}

func statePlot(w *ui.Window) {
	w.Row(25).Dynamic(2)

	change := false
	change = change || w.PropertyFloat("Initial investment (USD):", 0, &sim.Param.cash, 50000.0, 10, 10, 10)
	change = change || w.PropertyFloat("Average sales per week:", 0, &sim.Param.weeklySales, 10000.0, 1, 0.2, 3)
	change = change || w.PropertyFloat("Cost of each unit (USD):", 0, &sim.Param.unitCost, 1000.0, 1, 0.2, 3)
	change = change || w.PropertyFloat("Margin for each unit (USD):", 0, &sim.Param.unitBenefit, 1000.0, 1, 0.2, 3)
	change = change || w.PropertyInt("Size of each shipment:", 1, &sim.Param.batch, 1000, 1, 1)
	change = change || w.PropertyInt("Shipment duration (days):", 1, &sim.Param.shipmentDelay, 100, 1, 1)
	change = change || w.PropertyInt("Simulation duration (days):", 1, &sim.Param.duration, 10000, 1, 1)
	change = change || w.PropertyFloat("Monthly storage per unit (USD):", 0, &sim.Param.unitMonthlyStorage, 100, 1, 0.2, 3)
	if change {
		sim = NewSimulation(sim.Param)
	}
	w.Row(0).Dynamic(1)
	img, err := sim.Plot()
	if err != nil {
		w.LabelWrap(err.Error())
	} else {
		plot := image.NewRGBA(img.Bounds())
		draw.Draw(plot, img.Bounds(), img, image.Point{}, draw.Src)
		w.Image(plot)
	}
}

func updatefn(w *ui.Window) {
	statePlot(w)
}

func main() {
	var param Parameters

	param.cash = 1000.0
	param.batch = 20
	param.unitCost = 25.0
	param.unitBenefit = 10.0
	param.weeklySales = 7.0
	param.shipmentDelay = 14
	param.duration = 365
	toPrint := false

	flag.Float64Var(&param.cash, "cash", param.cash, "initial investment (USD)")
	flag.Float64Var(&param.weeklySales, "sales", param.weeklySales, "average sales per week (quantity)")
	flag.Float64Var(&param.unitCost, "cost", param.unitCost, "cost of each unit (USD)")
	flag.Float64Var(&param.unitBenefit, "margin", param.unitBenefit, "margin for each unit (USD)")
	flag.IntVar(&param.batch, "batch", param.batch, "size of each shipment (quantity)")
	flag.IntVar(&param.shipmentDelay, "delay", param.shipmentDelay, "time to ship a batch (days)")
	flag.IntVar(&param.duration, "days", param.duration, "simulation duration (days)")
	flag.BoolVar(&toPrint, "print", toPrint, "output CSV values")
	flag.Parse()

	if toPrint {
		NewSimulation(param).Print()
	} else {
		sim = NewSimulation(param)
		wnd := ui.NewMasterWindow(0, "Sales Simulation", updatefn)
		wnd.SetStyle(style.FromTheme(style.DarkTheme, 2.0))
		wnd.Main()
	}
}

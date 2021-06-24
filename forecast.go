package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"os"
	"os/user"
	"path"
	"time"

	ui "github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/style"
	"github.com/go-yaml/yaml"
	"github.com/wcharczuk/go-chart"
)

type Shipment struct {
	arrival  int
	quantity int
}

type Parameters struct {
	Cash               float64
	BatchSize          int
	UnitCost           float64
	UnitBenefit        float64
	unitMonthlyStorage float64
	WeeklySales        float64
	ShipmentDelay      int
	InitialStock       int
	SimulationDuration int
}

type Config struct {
	Parameters
}

type Simulation struct {
	Date  []time.Time
	Stock []float64
	Cash  []float64
	Param *Parameters
}

var (
	sim Simulation

	configFileName = "forecast-config.yaml"
)

func NewSimulation(param *Parameters) Simulation {
	stock := float64(param.InitialStock)
	cash := param.Cash
	shipments := []Shipment{}

	duration := param.SimulationDuration * 30
	sim := Simulation{
		Date:  make([]time.Time, duration),
		Stock: make([]float64, duration),
		Cash:  make([]float64, duration),
		Param: param,
	}

	for day := 0; day < duration; day++ {
		date := time.Now().AddDate(0, 0, day)

		sim.Date[day] = date
		sim.Stock[day] = stock
		sim.Cash[day] = cash

		sellRate := param.WeeklySales / 7.0
		batchCost := float64(param.BatchSize) * param.UnitCost

		// storage cost
		unitDailyStorage := param.unitMonthlyStorage / 30
		cash -= stock * unitDailyStorage

		// if we can buy some more stock
		for cash >= batchCost {

			// optimal stock: two time the shipment delay
			runway := sellRate * float64(param.ShipmentDelay) * 2

			// compute the pending amount before buying
			// some more.
			pending := 0
			for _, shipment := range shipments {
				pending += shipment.quantity
			}

			// buy stock if we need some more
			if stock+float64(pending) < runway {
				shipments = append(shipments, Shipment{
					arrival:  day + param.ShipmentDelay,
					quantity: param.BatchSize,
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
			unitGain := param.UnitCost + param.UnitBenefit
			cash += sellRate * unitGain
		} else {
			unitGain := param.UnitCost + param.UnitBenefit
			cash += stock * unitGain
			stock = 0
		}
	}
	return sim
}

func (s Simulation) Print() {
	fmt.Printf("cash\t%.2f\tinitial investment\n", s.Param.Cash)
	fmt.Printf("sales\t%.2f\tweekly sales\n", s.Param.WeeklySales)
	fmt.Printf("storage\t%.2f\tstorage cost per unit per month\n", s.Param.unitMonthlyStorage)
	fmt.Printf("cost\t%.2f\tprice of each unit\n", s.Param.UnitCost)
	fmt.Printf("margin\t%.2f\tmargin for each unit\n", s.Param.UnitBenefit)
	fmt.Printf("delay\t%d\tdays to ship\n", s.Param.ShipmentDelay)
	fmt.Printf("\n")
	fmt.Printf("day\tcash\tstock\n")
	for i := range s.Date {
		fmt.Printf("%v\t%.2f\t%d\n", s.Date[i].Format("01-02-2006"), s.Cash[i], int(s.Stock[i]))
	}
}

func (s Simulation) Plot(width, height int) (image.Image, error) {

	stockStyle := chart.Style{
		StrokeColor: chart.GetDefaultColor(1),
		StrokeWidth: 2.0,
	}
	cashStyle := chart.Style{
		StrokeColor: chart.GetDefaultColor(2),
		StrokeWidth: 2.0,
	}
	graph := chart.Chart{
		Width:  width,
		Height: height,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    12,
				Bottom: 12,
				Left:   12,
				Right:  12,
			},
		},
		YAxis: chart.YAxis{
			Name: "Euro",
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
				Name:    "Stock available",
				Style:   stockStyle,
				XValues: s.Date,
				YValues: s.Stock,
				YAxis:   chart.YAxisSecondary,
			},
			chart.TimeSeries{
				Name:    "Cash available",
				Style:   cashStyle,
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
	w.Row(20).Dynamic(2)

	change := false
	change = change || w.PropertyFloat("Initial investment (Euro):", 0, &sim.Param.Cash, 50000.0, 10, 10, 10)
	change = change || w.PropertyFloat("Average sales per week:", 0, &sim.Param.WeeklySales, 1000.0, 1, 0.2, 3)
	change = change || w.PropertyFloat("Cost of each unit (Euro):", 0, &sim.Param.UnitCost, 2000.0, 1, 0.2, 3)
	change = change || w.PropertyFloat("Margin for each unit (Euro):", 0, &sim.Param.UnitBenefit, 1000.0, 1, 0.2, 3)
	change = change || w.PropertyInt("Size of each shipment:", 1, &sim.Param.BatchSize, 2000, 1, 1)
	change = change || w.PropertyInt("Shipment duration (days):", 1, &sim.Param.ShipmentDelay, 120, 1, 1)
	change = change || w.PropertyInt("Simulation duration (months):", 1, &sim.Param.SimulationDuration, 60, 1, 1)
	change = change || w.PropertyFloat("Monthly storage per unit (Euro):", 0, &sim.Param.unitMonthlyStorage, 100, 1, 0.2, 3)
	if change {
		sim = NewSimulation(sim.Param)
	}
	w.Row(0).Dynamic(1)
	rect := w.WidgetBounds()
	img, err := sim.Plot(rect.W, rect.H)
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

func readParams(param *Parameters) error {
	f, err := os.Open(configFileName)
	if err != nil {
		param.Cash = 1000.0
		param.BatchSize = 20
		param.UnitCost = 25.0
		param.UnitBenefit = 10.0
		param.WeeklySales = 7.0
		param.ShipmentDelay = 14
		param.SimulationDuration = 12
		return nil
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(param)
	if err != nil {
		return err
	}
	return nil
}

func writeParams(param *Parameters) error {
	f, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(param)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	usr, _ := user.Current()
	configFileName = path.Join(usr.HomeDir, configFileName)

	var toPrint = false
	var param Parameters
	if err := readParams(&param); err != nil {
		fmt.Printf("Failed to read config file: %w\n", err)
		os.Exit(1)
	}

	flag.Float64Var(&param.Cash, "cash", param.Cash, "initial investment (Euro)")
	flag.Float64Var(&param.WeeklySales, "sales", param.WeeklySales, "average sales per week (quantity)")
	flag.Float64Var(&param.UnitCost, "cost", param.UnitCost, "cost of each unit (Euro)")
	flag.Float64Var(&param.UnitBenefit, "margin", param.UnitBenefit, "margin for each unit (Euro)")
	flag.IntVar(&param.BatchSize, "batch", param.BatchSize, "size of each shipment (quantity)")
	flag.IntVar(&param.ShipmentDelay, "delay", param.ShipmentDelay, "time to ship a batch (days)")
	flag.IntVar(&param.SimulationDuration, "months", param.SimulationDuration, "simulation duration (months)")
	flag.BoolVar(&toPrint, "print", toPrint, "output CSV values")
	flag.StringVar(&configFileName, "config", configFileName, "config file")
	flag.Parse()

	if toPrint {
		NewSimulation(&param).Print()
	} else {
		sim = NewSimulation(&param)
		wnd := ui.NewMasterWindow(0, "Sales Simulation", updatefn)
		wnd.SetStyle(style.FromTheme(style.DarkTheme, 2.0))
		wnd.OnClose(func() {
			if err := writeParams(&param); err != nil {
				fmt.Printf("Failed to write config: %w\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		})
		wnd.Main()
	}
	writeParams(&param)
}

package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/jabgibson/fortress"
	"github.com/jabgibson/tomlseq"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
)

func main() {
	var flagSchematic string
	var flagHTTPGet bool
	var flagLog string
	flag.StringVar(&flagSchematic, "s", "", "")
	flag.BoolVar(&flagHTTPGet, "url", false, "")
	flag.StringVar(&flagLog, "l", "fortress-err.log", "")
	flag.Parse()

	l, err := zap.NewProduction()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	if flagSchematic == "" {
		l.Error("Empty schematic: A toml formatted schematic is required for fortress to run. ", zap.String("example", "-s=app.toml"))
		os.Exit(1)
	}

	var sc schematic
	var scBytes []byte
	if flagHTTPGet {
		rs, err := http.Get(flagSchematic)
		if err != nil {
			l.Error("Failed to get schematic from defined url", zap.String("schematic-url", flagSchematic))
			os.Exit(1)
		}
		defer rs.Body.Close()
		scBytes, err = ioutil.ReadAll(rs.Body)
		if err != nil {
			l.Error("Failed to read body of schematic from configured url", zap.String("url", flagSchematic))
			os.Exit(1)
		}
	} else {
		scBytes, err = ioutil.ReadFile(flagSchematic)
		if err != nil {
			l.Error("Failed to read body of schematic from file", zap.String("file", flagSchematic))
			os.Exit(1)
		}
	}
	sequenced := tomlseq.Process("_seq", scBytes)
	if _, err := toml.Decode(string(sequenced), &sc); err != nil {
		l.Error("Unable to decode schematic: Is the schematic defined correctly?", zap.String("s", flagSchematic), zap.ByteString("schematic", sequenced))
		os.Exit(1)
	}

	var orders []fortress.Orderer
	for _, eo := range sc.EnvOrders {
		orders = append(orders, eo)
	}
	for _, ro := range sc.RunOrders {
		orders = append(orders, ro)
	}
	for _, so := range sc.ScriptOrders {
		orders = append(orders, so)
	}
	for _, do := range sc.DataOrders {
		orders = append(orders, do)
	}
	sort.Sort(fortress.BySequence(orders))

	contexts := map[string]fortress.OrderContext{}
	globalContext := fortress.OrderContext{
		EnvVars: map[string]string{},
		Owner:   "#GLOBAL#",
		Data:    map[string]string{},
	}

	for _, order := range orders {
		contexts[order.Self().ID] = fortress.OrderContext{
			Owner:   order.Self().ID,
			EnvVars: map[string]string{},
		}
	}

	for _, order := range orders {
		// Before exeucting order, merge the global and order specific contexts unless configured to ignore
		var context fortress.OrderContext
		if order.Self().IgnoreGlobal {
			context = contexts[order.Self().ID]
		} else {
			context = mergeContexts(globalContext, contexts[order.Self().ID])
		}

		// execute order and capture report
		report := order.ExecuteOrder(context)

		// Env: For each EnvDirection in report, distribute them to the nessessary contexts
		for _, ed := range report.EnvDirections {
			// If global context, set the global context
			if ed.Target == "#GLOBAL#" {
				globalContext.EnvVars[ed.Key] = ed.Value
			} else {
				// If EnvDirection is targeting specific orders, update their contexts with EnvDirection details
				_, ok := contexts[ed.Target]
				if !ok {
					l.Warn("Trying to set environement variable to context with no owner. Skipping", zap.String("target", ed.Target))
					continue
				}
				contexts[ed.Target].EnvVars[ed.Key] = ed.Value
			}
		}

		// Script:

		// Data:
		for k, v := range report.Data {
			globalContext.Data[k] = v
		}

	}
}

func mergeContexts(global, specific fortress.OrderContext) (context fortress.OrderContext) {
	context.Owner = specific.Owner
	context.EnvVars = map[string]string{}
	context.Data = map[string]string{}
	for k, v := range specific.EnvVars {
		context.EnvVars[k] = v
	}

	// Propagate global environment vars unless variable already exists from order specific context
	for k, v := range global.EnvVars {
		if _, ok := context.EnvVars[k]; !ok {
			context.EnvVars[k] = v
		}
	}
	// propagate Shared Data to context
	for k, v := range global.Data {
		context.Data[k] = v
	}

	return context
}

const portOpen = 0
const portLocked = 1

func porter(o order) int {
	if serv, err := net.Listen("tcp", ":"+strconv.Itoa(o.Port)); err != nil {
		return portLocked
	} else {
		serv.Close()
		return portOpen
	}
}

type schematic struct {
	EnvOrders    []fortress.EnvOrder    `toml:"env"`
	RunOrders    []fortress.RunOrder    `toml:"run"`
	ScriptOrders []fortress.ScriptOrder `toml:"script"`
	DataOrders   []fortress.DataOrder   `toml:"data"`
}

type order struct {
	Name string   `toml:"name"`
	Port int      `toml:"port"`
	Run  string   `toml:"run"`
	Args []string `toml:"args"`
}

package main

import (
	"flag"
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"github.com/tealeg/xlsx"
	"os"
	"regexp"
	"strconv"
	"time"
)

// flag
var (
	input = flag.String(
		"input",
		"",
		"input anno txt",
	)
	output = flag.String(
		"output",
		"",
		"output xlsx name",
	)
)

// regexp
var (
	isHgmd    = regexp.MustCompile("DM")
	isClinvar = regexp.MustCompile("Pathogenic|Likely_pathogenic")
	//indexReg   = regexp.MustCompile(`\d+\.\s+`)
	//newlineReg = regexp.MustCompile(`\n+`)
)

// Tier1 >0
// LoF 2
var FuncInfo = map[string]int{
	"splice-3":   2,
	"splice-5":   2,
	"inti-loss":  2,
	"alt-start":  2,
	"frameshift": 2,
	"nonsense":   2,
	"stop-gain":  2,
	"span":       2,
	"missense":   1,
	"cds-del":    1,
	"cds-indel":  1,
	"cds-ins":    1,
	"splice-10":  1,
	"splice+10":  1,
}

var AFlist = []string{
	"1000G ASN AF",
	"1000G EAS AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC EAS AF",
	"ExAC AF",
	"PVFD AF",
	"Panel AlleleFreq",
}

func main() {
	t0 := time.Now()

	flag.Parse()
	if *input == "" || *output == "" {
		flag.Usage()
		os.Exit(0)
	}
	data, title := simple_util.File2MapArray(*input, "\t")
	title = append(title, "Tier")
	t1 := time.Now()
	fmt.Printf("load anno file took  %v to run.\n", t1.Sub(t0))

	var stats = make(map[string]int)
	outputXlsx := xlsx.NewFile()
	outputSheet, err := outputXlsx.AddSheet("filter_variant")
	simple_util.CheckErr(err)

	var outputRow = outputSheet.AddRow()

	for _, str := range title {
		outputCell := outputRow.AddCell()
		outputCell.SetString(str)
	}

	stats["Total"] = len(data)
	for _, item := range data {
		if isHgmd.MatchString(item["HGMD Pred"]) || isClinvar.MatchString(item["ClinVar Significance"]) {
			stats["HGMD/ClinVar"]++
			if checkAF(item, 0.01) {
				item["Tier"] = "Tier1"
				stats["HGMD/ClinVar Tier1"]++
			} else {
				item["Tier"] = "Tier2"
				stats["HGMD/ClinVar Tier2"]++
			}
		} else if item["ACMG"] == "Benign" || item["ACMG"] == "Likely Benign" {
			stats["B/LB"]++
			continue
		}

		if checkAF(item, 0.01) {
			stats["noB/LB AF<=0.01"]++
			if FuncInfo[item["Function"]] > 0 {
				item["Tier"] = "Tier1"
				stats["noB/LB Tier1"]++
			} else if item["Tier"] != "Tier1" {
				item["Tier"] = "Tier2"
				stats["noB/LB Tier2"]++
			}
		} else if item["Tier"] != "Tier1" {
			item["Tier"] = "Tier2"
			stats["noB/LB Tier2"]++
		}

		if item["Tier"] == "Tier1" {
			stats["Tier1"]++
		}

		stats["Keep"]++
		outputRow = outputSheet.AddRow()
		for _, str := range title {
			outputCell := outputRow.AddCell()
			outputCell.SetString(item[str])
		}
	}
	fmt.Printf("Total        Variant : %d\n", stats["Total"])
	fmt.Printf("HGMD/ClinVar Hit     : %d\n", stats["HGMD/ClinVar"])
	fmt.Printf("HGMD/ClinVar Tier1   : %d\n", stats["HGMD/ClinVar Tier1"])
	fmt.Printf("B/LB         Skip    : %d\n", stats["B/LB"])
	fmt.Printf("no B/LB   AF<=0.01   : %d\n", stats["noB/LB AF<=0.01"])
	fmt.Printf("no B/LB      Tier1   : %d\n", stats["noB/LB Tier1"])
	fmt.Printf("no B/LB      Tier2   : %d\n", stats["noB/LB Tier2"])
	fmt.Printf("Keep         Variant : %d\n", stats["Keep"])
	fmt.Printf("Keep Tier1   Variant : %d\n", stats["Tier1"])
	t2 := time.Now()
	fmt.Printf("create excel took    %v to run.\n", t2.Sub(t1))

	err = outputXlsx.Save(*output)
	simple_util.CheckErr(err)
	t3 := time.Now()
	fmt.Printf("save excel file took %v to run.\n", t3.Sub(t2))
	fmt.Printf("total work took      %v to run.\n", t3.Sub(t0))
}

func checkAF(item map[string]string, threshold float64) bool {
	for _, key := range AFlist {
		af := item[key]
		if af == "" || af == "." {
			continue
		}
		AF, err := strconv.ParseFloat(af, 64)
		simple_util.CheckErr(err)
		if AF > threshold {
			return false
		}
	}
	return true
}

package anno

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/stringsUtil"
)

var long2short = map[string]string{
	"Pathogenic":             "P",
	"Likely Pathogenic":      "LP",
	"Uncertain Significance": "VUS",
	"Likely Benign":          "LB",
	"Benign":                 "B",
	"P":                      "P",
	"LP":                     "LP",
	"VUS":                    "VUS",
	"LB":                     "LB",
	"B":                      "B",
}

// regexp
var (
	indexReg = regexp.MustCompile(`\d+\.\s+`)

	inBrackets = regexp.MustCompile(`\(\S+\)`)

	rmChr = regexp.MustCompile(`^chr`)
	cds   = regexp.MustCompile(`^C`)
	ratio = regexp.MustCompile(`^[01](.\d+)?$`)
	reInt = regexp.MustCompile(`^\d+$`)

	isARorXR = regexp.MustCompile(`AR|XR`)
	isAR     = regexp.MustCompile(`AR`)
	isAD     = regexp.MustCompile(`AD`)
	isXL     = regexp.MustCompile(`XL`)
	isYL     = regexp.MustCompile(`YL`)

	withHom = regexp.MustCompile(`Hom`)

	isHet  = regexp.MustCompile(`^Het`)
	isHom  = regexp.MustCompile(`^Hom`)
	isHemi = regexp.MustCompile(`^Hemi`)
	isNA   = regexp.MustCompile(`^NA`)

	isHetNA = regexp.MustCompile(`^Het;NA`)
	isNAHet = regexp.MustCompile(`^NA;Het`)

	isHomHetHet = regexp.MustCompile(`^Hom;Het;Het`)
	isHomNANA   = regexp.MustCompile(`^Hom;NA;NA`)

	isHetHetHet = regexp.MustCompile(`^Het;Het;Het`)
	isHetHetNA  = regexp.MustCompile(`^Het;Het;NA`)
	isHetNAHet  = regexp.MustCompile(`^Het;NA;Het`)
	isHetNANA   = regexp.MustCompile(`^Het;NA;NA`)

	isHomInherit      = regexp.MustCompile(`^Hom;Het;Het|^Hom;Het;NA|^Hom;NA;Het|^Hom;NA;NA`)
	isXLInheritMale   = regexp.MustCompile(`^Hemi;Het;NA|^Hemi;NA;Het|^Hemi;NA;NA|^Het;Het;NA|^Het;NA;Het|^Het;NA;NA`)
	isXLInheritFemale = regexp.MustCompile(`^Hom;Het;NA|^Hom;NA;Het|^Hom;NA;NA|^Het;NA;NA`)
	isYLInherit       = regexp.MustCompile(`^Hemi;NA;NA|^Hemi;Hom;NA|^Hemi;Het;NA|^Hemi;NA;Hom|^Hemi;NA;Het|^Het;Hom;NA|^Het;Het;NA|^Het;NA;Hom|^Het;NA;Het|^Het;NA;NA`)
)

//UpdateSnvTier1 add other info for tier1 variant
func UpdateSnvTier1(item map[string]string) {

	item["一键搜索链接"] = url.QueryEscape(googleKey(item))
	item["cHGVS"] = cHgvsAlt(item["cHGVS"])

	// addition
	item["烈性突变"] = "否"
	if FuncInfo[item["Function"]] == 3 {
		item["烈性突变"] = "是"
	}

	item["HGMDorClinvar"] = "否"
	if isHgmd.MatchString(item["HGMD Pred"]) {
		item["HGMDorClinvar"] = "是"
	}
	if isClinvar.MatchString(item["ClinVar Significance"]) {
		item["HGMDorClinvar"] = "是"
	}

	item["GnomAD homo"] = item["GnomAD HomoAlt Count"]
	item["GnomAD hemi"] = item["GnomAD HemiAlt Count"]
	item["纯合，半合"] = item["GnomAD HomoAlt Count"] // + "|" + dataHash["GnomAD HemiAlt Count"]

	item["历史样本检出个数"] = item["sampleMut"] + "/" + item["sampleAll"]

	// remove index
	for _, k := range [2]string{"GeneralizationEN", "GeneralizationCH"} {
		sep := "\n\n"
		key := strings.Split(item[k], sep)
		for i := range key {
			key[i] = indexReg.ReplaceAllLiteralString(key[i], "")
		}
		item[k] = strings.Join(key, sep)
	}

}

//Score2Pred add _pred for scores
func Score2Pred(item map[string]string) {
	score, e := strconv.ParseFloat(item["dbscSNV_ADA_SCORE"], 32)
	if e != nil {
		item["dbscSNV_ADA_pred"] = item["dbscSNV_ADA_SCORE"]
	} else {
		if score >= 0.6 {
			item["dbscSNV_ADA_pred"] = "D"
		} else {
			item["dbscSNV_ADA_pred"] = "P"
		}
	}
	score, e = strconv.ParseFloat(item["dbscSNV_RF_SCORE"], 32)
	if e != nil {
		item["dbscSNV_RF_pred"] = item["dbscSNV_RF_SCORE"]
	} else {
		if score >= 0.6 {
			item["dbscSNV_RF_pred"] = "D"
		} else {
			item["dbscSNV_RF_pred"] = "P"
		}
	}

	// ＞=2.0 保守
	score, e = strconv.ParseFloat(item["GERP++_RS"], 32)
	if e != nil {
		item["GERP++_RS_pred"] = item["GERP++_RS"]
	} else {
		if score >= 2.0 {
			item["GERP++_RS_pred"] = "保守"
		} else {
			item["GERP++_RS_pred"] = "不保守"
		}
	}
	score, e = strconv.ParseFloat(item["PhyloP Vertebrates"], 32)
	if e != nil {
		item["PhyloP Vertebrates Pred"] = item["PhyloP Vertebrates"]
	} else {
		if score >= 2.0 {
			item["PhyloP Vertebrates Pred"] = "保守"
		} else {
			item["PhyloP Vertebrates Pred"] = "不保守"
		}
	}
	score, e = strconv.ParseFloat(item["PhyloP Placental Mammals"], 32)
	if e != nil {
		item["PhyloP Placental Mammals Pred"] = item["PhyloP Placental Mammals"]
	} else {
		if score >= 2.0 {
			item["PhyloP Placental Mammals Pred"] = "保守"
		} else {
			item["PhyloP Placental Mammals Pred"] = "不保守"
		}
	}
}

var xparReg = [][]int{
	{60000, 2699520},
	{154931043, 155260560},
}
var yparReg = [][]int{
	{10000, 2649520},
	{59034049, 59363566},
}
var (
	isChrX  = regexp.MustCompile(`X`)
	isChrY  = regexp.MustCompile(`Y`)
	isChrXY = regexp.MustCompile(`[XY]`)
	isMale  = regexp.MustCompile(`M`)
	isDel   = regexp.MustCompile(`del`)
)

func inPAR(chr string, start, end int) bool {
	if isChrX.MatchString(chr) {
		for _, par := range xparReg {
			if start < par[1] && end > par[0] {
				return true
			}
		}
	} else if isChrY.MatchString(chr) {
		for _, par := range yparReg {
			if start < par[1] && end > par[0] {
				return true
			}
		}
	}
	return false
}

//UpdateSnv add info for all variant
func UpdateSnv(item map[string]string, gender string, debug bool) {

	// #Chr+Stop
	item["#Chr"] = "chr" + rmChr.ReplaceAllString(item["#Chr"], "")
	if item["VarType"] == "snv" {
		item["#Chr+Stop"] = item["#Chr"] + ":" + item["Stop"]
		item["chr-show"] = item["#Chr"] + ":" + item["Stop"]
	} else {
		item["#Chr+Stop"] = item["#Chr"] + ":" + item["Start"] + "-" + item["Stop"]
		if isDel.MatchString(item["VarType"]) {
			item["chr-show"] = item["#Chr"] + ":" + stringsUtil.StringPlus(item["Start"], 1) + ".." + item["Stop"]
		} else {
			item["chr-show"] = item["#Chr"] + ":" + item["Start"] + ".." + stringsUtil.StringPlus(item["Stop"], 1)
		}
	}

	// pHGVS= pHGVS1+"|"+pHGVS3
	if item["pHGVS1"] != "" && item["pHGVS3"] != "" && item["pHGVS1"] != "." && item["pHGVS3"] != "." {
		item["pHGVS"] = item["pHGVS1"] + " | " + item["pHGVS3"]
	}

	MutationNameArray := strings.Split(item["MutationName"], ":")
	if len(MutationNameArray) > 1 {
		item["MutationNameLite"] = inBrackets.ReplaceAllString(MutationNameArray[0], "") + ":" + MutationNameArray[1]
		//item["MutationNameLite"] = item["Transcript"] + ":" + strings.Split(item["MutationName"], ":")[1]
	} else {
		item["MutationNameLite"] = item["MutationName"]
	}

	// Zygosity format
	item["Zygosity"] = zygosityFormat(item["Zygosity"])

	chr := item["#Chr"]
	if isChrXY.MatchString(chr) && isMale.MatchString(gender) {
		start, e := strconv.Atoi(item["Start"])
		simpleUtil.CheckErr(e)
		stop, e := strconv.Atoi(item["Stop"])
		simpleUtil.CheckErr(e)
		if !inPAR(chr, start, stop) && withHom.MatchString(item["Zygosity"]) {
			zygosity := strings.Split(item["Zygosity"], ";")
			genders := strings.Split(gender, ",")
			if len(genders) <= len(zygosity) {
				for i := range genders {
					if isMale.MatchString(genders[i]) && isHom.MatchString(zygosity[i]) {
						zygosity[i] = strings.Replace(zygosity[i], "Hom", "Hemi", 1)
					}
				}
				item["Zygosity"] = strings.Join(zygosity, ";")
			} else {
				log.Fatalf("conflict gender[%s]and Zygosity[%s]\n", gender, item["Zygosity"])
			}
		}
	}

	if debug && item["自动化判断"] != long2short[item["ACMG"]] {
		_, err = fmt.Fprintf(
			os.Stderr,
			"acmg conflict:[%s=>%s]:%s\n",
			long2short[item["ACMG"]], item["自动化判断"], item["MutationName"],
		)
		simpleUtil.CheckErr(err)
	}
	return
}

//InheritCheck count variants of gene
func InheritCheck(item map[string]string, inheritDb map[string]map[string]int) {
	geneSymbol := item["Gene Symbol"]
	inherit := item["ModeInheritance"]
	zygosity := item["Zygosity"]
	var db = make(map[string]int)
	if inheritDb[geneSymbol] == nil {
		inheritDb[geneSymbol] = db
	}
	if isARorXR.MatchString(inherit) {
		if isHet.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag1"]++
		}
		if isHetNA.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag10"]++
		}
		if isNAHet.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag01"]++
		}
		if isHetHetNA.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag110"]++
		}
		if isHetNAHet.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag101"]++
		}
		if isHetNANA.MatchString(zygosity) {
			inheritDb[geneSymbol]["flag100"]++
		}
	}
}

//InheritCoincide calculate 遗传相符
func InheritCoincide(item map[string]string, inheritDb map[string]map[string]int, isTrio bool) string {
	geneSymbol := item["Gene Symbol"]
	inherit := item["ModeInheritance"]
	zygosity := item["Zygosity"]
	if isTrio {
		if isNA.MatchString(zygosity) {
			return "NA"
		}
		if isAD.MatchString(inherit) &&
			(isHetNANA.MatchString(zygosity) || isHomNANA.MatchString(zygosity)) {
			return "相符"
		}
		if isXL.MatchString(inherit) && (isXLInheritMale.MatchString(zygosity) || isXLInheritFemale.MatchString(zygosity)) {
			return "相符"
		}
		if isYL.MatchString(inherit) && isYLInherit.MatchString(zygosity) {
			return "相符"
		}
		if isAR.MatchString(inherit) {
			if isHomInherit.MatchString(zygosity) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag110"] > 0 &&
				inheritDb[geneSymbol]["flag101"] > 0 &&
				(isHetHetNA.MatchString(zygosity) || isHetNAHet.MatchString(zygosity)) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag110"] > 0 &&
				inheritDb[geneSymbol]["flag100"] > 0 &&
				(isHetHetNA.MatchString(zygosity) || isHetNANA.MatchString(zygosity)) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag101"] > 0 &&
				inheritDb[geneSymbol]["flag100"] > 0 &&
				(isHetNAHet.MatchString(zygosity) || isHetNANA.MatchString(zygosity)) {
				return "相符"
			}
			if inheritDb[geneSymbol]["flag100"] >= 2 && isHetNANA.MatchString(zygosity) {
				return "相符"
			}
		}
		if isAD.MatchString(inherit) {
			if isHetHetNA.MatchString(zygosity) || isHetNAHet.MatchString(zygosity) {
				return "不确定"
			}
		}
		if isAR.MatchString(inherit) {
			if isHetHetHet.MatchString(zygosity) || isHetHetNA.MatchString(zygosity) ||
				isHetNAHet.MatchString(zygosity) || isHetNANA.MatchString(zygosity) {
				return "不确定"
			}
		}
		return "不相符"
	} else {
		if isXL.MatchString(inherit) || isYL.MatchString(inherit) {
			if isHet.MatchString(zygosity) || isHom.MatchString(zygosity) || isHemi.MatchString(zygosity) {
				return "相符"
			}
		}
		if isAD.MatchString(inherit) {
			if isHet.MatchString(zygosity) || isHom.MatchString(zygosity) {
				return "相符"
			}
		}
		if isAR.MatchString(inherit) {
			if isHom.MatchString(zygosity) {
				return "相符"
			}
			if isHet.MatchString(zygosity) {
				if inheritDb[geneSymbol]["flag1"] >= 2 {
					return "相符"
				} else {
					return "不确定"
				}
			}
		}
		return "不相符"
	}
}

// FamilyTag return familyTag
func FamilyTag(item map[string]string, inheritDb map[string]map[string]int, tag string) string {
	geneSymbol := item["Gene Symbol"]
	inherit := item["ModeInheritance"]
	zygosity := item["Zygosity"]
	if tag == "couple" {
		if isARorXR.MatchString(inherit) {
			if inheritDb[geneSymbol]["flag10"] > 0 &&
				inheritDb[geneSymbol]["flag01"] > 0 &&
				(isHetNA.MatchString(zygosity) || isNAHet.MatchString(zygosity)) {
				return "couple-CP"
			}
		}
	} else if tag == "trio" {
		if isAD.MatchString(inherit) && isHetNANA.MatchString(zygosity) {
			return "trio-AD"
		}
		if isARorXR.MatchString(inherit) {
			if inheritDb[geneSymbol]["flag110"] > 0 &&
				inheritDb[geneSymbol]["flag101"] > 0 &&
				(isHetHetNA.MatchString(zygosity) || isHetNAHet.MatchString(zygosity)) {
				return "trio-CP"
			}
		}
		if isAR.MatchString(inherit) && isHomHetHet.MatchString(zygosity) {
			return "trio-hom"
		}
	} else {
		if isAR.MatchString(inherit) && isHom.MatchString(zygosity) {
			return "AR-Hom"
		}
		if isAR.MatchString(inherit) && inheritDb[geneSymbol]["flag1"] > 1 && isHet.MatchString(zygosity) {
			return "AR-CP"
		}
		if isXL.MatchString(inherit) && (isHom.MatchString(zygosity) || isHemi.MatchString(zygosity)) {
			return "XL-Hom/Hemi"
		}
	}
	return ""
}

func zygosityFormat(zygosity string) string {
	zygosity = strings.Replace(zygosity, "het-ref", "Het", -1)
	zygosity = strings.Replace(zygosity, "het-alt", "Het", -1)
	zygosity = strings.Replace(zygosity, "hom-alt", "Hom", -1)
	zygosity = strings.Replace(zygosity, "hem-alt", "Hemi", -1)
	zygosity = strings.Replace(zygosity, "hemi-alt", "Hemi", -1)
	return zygosity
}

var inheritFromMap = map[string]string{
	"Het":    "（杂合）",
	"Hom":    "（纯合）",
	"Hemi":   "（半合）",
	"UC":     "不确定",
	"Denovo": "新发",
	"NA":     "NA",
}

//InheritFrom return 变异来源
func InheritFrom(item map[string]string, sampleList []string) string {
	if len(sampleList) < 3 {
		return "NA1"
	}
	zygosity := item["Zygosity"]
	zygos := strings.Split(zygosity, ";")
	if len(zygos) < 3 {
		return "NA2"
	}
	zygos3 := strings.Join(zygos[0:3], ";")
	//fmt.Println(zygos3)
	var from string
	switch zygos3 {
	case "Hom;Hom;Hom":
		from = sampleList[1] + inheritFromMap["Hom"] + "/" + sampleList[2] + inheritFromMap["Hom"]
	case "Hom;Hom;Het":
		from = sampleList[1] + inheritFromMap["Hom"] + "/" + sampleList[2] + inheritFromMap["Het"]
	case "Hom;Hom;Hemi":
		from = sampleList[1] + inheritFromMap["Hom"] + "/" + sampleList[2] + inheritFromMap["Hemi"]
	case "Hom;Hom;NA":
		from = sampleList[1] + inheritFromMap["Hom"] + "/" + inheritFromMap["Denovo"]

	case "Hom;Het;Hom":
		from = sampleList[1] + inheritFromMap["Het"] + "/" + sampleList[2] + inheritFromMap["Hom"]
	case "Hom;Het;Het":
		from = sampleList[1] + inheritFromMap["Het"] + "/" + sampleList[2] + inheritFromMap["Het"]
	case "Hom;Het;Hemi":
		from = sampleList[1] + inheritFromMap["Het"] + "/" + sampleList[2] + inheritFromMap["Hemi"]
	case "Hom;Het;NA":
		from = sampleList[1] + inheritFromMap["Het"] + "/" + inheritFromMap["Denovo"]

	case "Hom;Hemi;Hom":
		from = sampleList[1] + inheritFromMap["Hemi"] + "/" + sampleList[2] + inheritFromMap["Hom"]
	case "Hom;Hemi;Het":
		from = sampleList[1] + inheritFromMap["Hemi"] + "/" + sampleList[2] + inheritFromMap["Het"]
	case "Hom;Hemi;Hemi":
		from = inheritFromMap["NA"]
	case "Hom;Hemi;NA":
		from = sampleList[1] + inheritFromMap["Hemi"] + "/" + inheritFromMap["Denovo"]

	case "Hom;NA;Hom":
		from = inheritFromMap["Denovo"] + "/" + sampleList[2] + inheritFromMap["Hom"]
	case "Hom;NA;Het":
		from = inheritFromMap["Denovo"] + "/" + sampleList[2] + inheritFromMap["Het"]
	case "Hom;NA;Hemi":
		from = inheritFromMap["Denovo"] + "/" + sampleList[2] + inheritFromMap["Hemi"]
	case "Hom;NA;NA":
		from = inheritFromMap["Denovo"]

	case "Het;Hom;Hom":
		from = inheritFromMap["UC"]
	case "Het;Hom;Het":
		from = inheritFromMap["UC"]
	case "Het;Hom;Hemi":
		from = inheritFromMap["UC"]
	case "Het;Hom;NA":
		from = sampleList[1] + inheritFromMap["Hom"]

	case "Het;Het;Hom":
		from = inheritFromMap["UC"]
	case "Het;Het;Het":
		from = inheritFromMap["UC"]
	case "Het;Het;Hemi":
		from = inheritFromMap["UC"]
	case "Het;Het;NA":
		from = sampleList[1] + inheritFromMap["Het"]

	case "Het;Hemi;Hom":
		from = inheritFromMap["UC"]
	case "Het;Hemi;Het":
		from = inheritFromMap["UC"]
	case "Het;Hemi;Hemi":
		from = inheritFromMap["NA"]
	case "Het;Hemi;NA":
		from = sampleList[1] + inheritFromMap["Hemi"]

	case "Het;NA;Hom":
		from = sampleList[2] + inheritFromMap["Hom"]
	case "Het;NA;Het":
		from = sampleList[2] + inheritFromMap["Het"]
	case "Het;NA;Hemi":
		from = sampleList[2] + inheritFromMap["Hemi"]
	case "Het;NA;NA":
		from = inheritFromMap["Denovo"]

	case "Hemi;Hom;Hom":
		from = inheritFromMap["UC"]
	case "Hemi;Hom;Het":
		from = inheritFromMap["UC"]
	case "Hemi;Hom;Hemi":
		from = sampleList[1] + inheritFromMap["Hom"]
	case "Hemi;Hom;NA":
		from = sampleList[1] + inheritFromMap["Hom"]

	case "Hemi;Het;Hom":
		from = inheritFromMap["UC"]
	case "Hemi;Het;Het":
		from = inheritFromMap["UC"]
	case "Hemi;Het;Hemi":
		from = sampleList[1] + inheritFromMap["Het"]
	case "Hemi;Het;NA":
		from = sampleList[1] + inheritFromMap["Het"]

	case "Hemi;Hemi;Hom":
		from = sampleList[2] + inheritFromMap["Hom"]
	case "Hemi;Hemi;Het":
		from = sampleList[2] + inheritFromMap["Het"]
	case "Hemi;Hemi;Hemi":
		from = inheritFromMap["NA"]
	case "Hemi;Hemi;NA":
		from = inheritFromMap["Denovo"]

	case "Hemi;NA;Hom":
		from = sampleList[2] + inheritFromMap["Hom"]
	case "Hemi;NA;Het":
		from = sampleList[2] + inheritFromMap["Het"]
	case "Hemi;NA;Hemi":
		from = inheritFromMap["Denovo"]
	case "Hemi;NA;NA":
		from = inheritFromMap["Denovo"]

	default:
		from = "NA3"
	}

	return from
}

var tr = map[rune]rune{
	'A': 'T',
	'C': 'G',
	'G': 'C',
	'T': 'A',
	'a': 't',
	'c': 'g',
	'g': 'c',
	't': 'a',
}

func reverseComplement(s string) string {
	runes := []rune(s)
	for i := range runes {
		if tr[runes[i]] != '\x00' {
			runes[i] = tr[runes[i]]
		}
	}
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

var err error

//PrimerDesign return 引物设计
func PrimerDesign(item map[string]string) string {
	var transcript = item["Transcript"]

	var pos string
	if item["VarType"] == "snv" {
		pos = item["Stop"]
	} else {
		pos = item["Start"] + "-" + item["Stop"]
	}
	var flank = item["Flank"]
	if item["Strand"] == "-" {
		flank = reverseComplement(flank)
	}
	funcRegion := cds.ReplaceAllString(item["FuncRegion"], "CDS")

	var Adepth int
	adepth := strings.Split(item["A.Depth"], ";")[0]
	if reInt.MatchString(adepth) {
		Adepth, err = strconv.Atoi(adepth)
		simpleUtil.CheckErr(err)
	}

	aratio := strings.Split(item["A.Ratio"], ";")[0]
	if ratio.MatchString(aratio) && aratio != "0" {
		Aratio, e := strconv.ParseFloat(aratio, 32)
		simpleUtil.CheckErr(e)

		aratio = strconv.FormatFloat(Aratio*100, 'f', 0, 32)
		if item["Depth"] == "" && Adepth > 0 {
			item["Depth"] = fmt.Sprintf("%.0f", float64(Adepth)/Aratio)
		}
	}

	primer := strings.Join(
		[]string{
			item["Gene Symbol"],
			transcript,
			item["cHGVS"],
			item["pHGVS3"],
			item["ExIn_ID"],
			funcRegion,
			strings.Split(item["Zygosity"], ";")[0],
			flank,
			item["exonCount"],
			item["Depth"],
			aratio,
			item["#Chr"], pos,
		}, "; ",
	)
	return primer
}

//exomePrimer return 引物设计 for exon cnv
func exomePrimer(item map[string]string) (primer string) {
	var transMap = make(map[string]string)
	var trans = strings.Split(item["Transcript"], ",")
	for i, k := range strings.Split(item["exons.hg19"], ",") {
		for _, gene := range strings.Split(k, "_") {
			transMap[gene] = trans[i]
		}
	}
	var genes = strings.Split(item["OMIM_Gene"], ";")
	var exons = strings.Split(item["OMIM_exon"], ";")
	var t string
	if item["type"] == "duplication" {
		t = "DUP"
	} else if item["type"] == "deletion" {
		t = "DEL"
	}
	var primers []string
	for i, gene := range genes {
		if gene == "" || gene == "-" {
			continue
		}
		primers = append(
			primers,
			strings.Join(
				[]string{
					gene, transMap[gene], exons[i] + " " + t, "-", exons[i], "-", "-", "-", "-", "-", "-", "-", "-",
				},
				";",
			),
		)
	}
	primer = strings.Join(primers, "\n")
	return
}

//largePrimer return 引物设计 for large cnv
func largePrimer(item map[string]string) (primer string) {
	summary := item["Summary"]
	infos := strings.SplitN(summary, "[", 2)
	primer = strings.Replace(infos[0], ",", "", -1)
	primer = strings.Replace(primer, "\"", "", -1)
	primer = strings.Join([]string{primer, "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}, ";")
	return
}

//CnvPrimer return 引物设计 for cnv
func CnvPrimer(item map[string]string, cnvType string) (primer string) {
	if cnvType == "exon_cnv" {
		primer = exomePrimer(item)
	} else if cnvType == "large_cnv" {
		primer = largePrimer(item)
	}
	return
}

// regexp
var (
	rsID     = regexp.MustCompile(`[rsRS]?\d+`)
	cHGVSalt = regexp.MustCompile(`alt: (\S+) \)`)
	cHGVS1   = regexp.MustCompile(`[cn]\.(\S+)(\S)>(\S)`)
	cHGVS2   = regexp.MustCompile(`[cn]\.(\S+)_(\S+)(del|ins)(\S+)`)
	cHGVS3   = regexp.MustCompile(`[cn]\.(\d+)(del|ins)(\S+)`)
	cHGVS4   = regexp.MustCompile(`[cn]\.(\d+[+-]\d+)(del|ins)(\S+)`)
	cHGVS5   = regexp.MustCompile(`[cn]\.(\S+)`)
	pHGVS1   = regexp.MustCompile(`p.\(=\) \(alt: p.(\S+) \)`)
	pHGVS2   = regexp.MustCompile(`p.\S+ \(std: p.\S+ alt: p.(\S+) \) \| p.\S+ \(std: p.\S+ alt: p.(\S+) \)`)
	pHGVS3   = regexp.MustCompile(`p.(\S+) \| p.(\S+)`)
	ivs1     = regexp.MustCompile(`c\.\d+([+-]\d+)(.*)$`)
	ivs2     = regexp.MustCompile(`c\.[-*]\d+([+-]\d+)(.*)$`)
	ivs3     = regexp.MustCompile(`c\.(\d+)([+-]\d+)_(\d+)([+-]\d+)(.*)$`)
	ivs4     = regexp.MustCompile(`c\.([-*]\d+)([+-]\d+)_([-*]\d+)([+-]\d+)(.*)$`)
)

func cHgvsAlt(cHgvs string) string {
	if m := cHGVSalt.FindStringSubmatch(cHgvs); m != nil {
		return m[1]
	}
	return cHgvs
}

func googleKey(item map[string]string) string {
	gene, chgvs, phgvs := item["Gene Symbol"], item["cHGVS"], item["pHGVS"]
	var searchKey []string

	// cHGVS
	var m []string
	if m = cHGVSalt.FindStringSubmatch(chgvs); m != nil {
		chgvs = m[1]
	}
	if m = cHGVS1.FindStringSubmatch(chgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s%s>%s", m[1], m[2], m[3]),
				fmt.Sprintf("%s%s->%s", m[1], m[2], m[3]),
				fmt.Sprintf("%s%s-->%s", m[1], m[2], m[3]),
				fmt.Sprintf("%s%s/%s", m[1], m[2], m[3]),
			)
	} else if m = cHGVS2.FindStringSubmatch(chgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s_%s%s%s", m[1], m[2], m[3], m[4]),
				fmt.Sprintf("%s_%s%s", m[1], m[2], m[3]),
				fmt.Sprintf("%s-%s%s%s", m[1], m[2], m[3], m[4]),
				fmt.Sprintf("%s-%s%s", m[1], m[2], m[3]),
			)
	} else if m = cHGVS3.FindStringSubmatch(chgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s%s%s", m[1], m[2], m[3]),
				fmt.Sprintf("%s%s", m[1], m[2]),
			)
	} else if m = cHGVS4.FindStringSubmatch(chgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s%s%s", m[1], m[2], m[3]),
				fmt.Sprintf("%s%s", m[1], m[2]),
			)
	} else if m = cHGVS5.FindStringSubmatch(chgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s", m[1]),
			)
	}

	// pHGVS
	if m = pHGVS1.FindStringSubmatch(phgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s", m[1]),
			)
	} else if m = pHGVS2.FindStringSubmatch(phgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s", m[1]),
				fmt.Sprintf("%s", m[2]),
			)
		if strings.Contains(m[2], "*") {
			searchKey =
				append(
					searchKey,
					strings.Replace(m[1], "*", "X", 1),
					strings.Replace(m[2], "*", "X", 1),
					strings.Replace(m[2], "*", "Ter", 1),
				)
		}
	} else if m = pHGVS3.FindStringSubmatch(phgvs); m != nil {
		searchKey =
			append(
				searchKey,
				fmt.Sprintf("%s", m[1]),
				fmt.Sprintf("%s", m[2]),
			)
		if strings.Contains(m[2], "*") {
			searchKey =
				append(
					searchKey,
					strings.Replace(m[1], "*", "X", 1),
					strings.Replace(m[2], "*", "X", 1),
					strings.Replace(m[2], "*", "Ter", 1),
				)
		}
	} else if strings.Contains(item["ExIn_ID"], "IVS") {
		intr := strings.Replace(item["ExIn_ID"], "IVS", "", 1)
		if m = ivs3.FindStringSubmatch(chgvs); m != nil {
			searchKey =
				append(
					searchKey,
					fmt.Sprintf("IVS %s%s_%s", intr, m[2], m[4]),
					fmt.Sprintf("IVS%s%s_%s", intr, m[2], m[4]),
				)
		} else if m = ivs4.FindStringSubmatch(chgvs); m != nil {
			searchKey =
				append(
					searchKey,
					fmt.Sprintf("IVS %s%s_%s", intr, m[2], m[4]),
					fmt.Sprintf("IVS%s%s_%s", intr, m[2], m[4]),
				)
		} else if m = ivs1.FindStringSubmatch(chgvs); m != nil {
			searchKey =
				append(
					searchKey,
					fmt.Sprintf("IVS %s%s", intr, m[1]),
					fmt.Sprintf("IVS%s%s", intr, m[1]),
				)
		} else if m = ivs2.FindStringSubmatch(chgvs); m != nil {
			searchKey =
				append(
					searchKey,
					fmt.Sprintf("IVS %s%s", intr, m[1]),
					fmt.Sprintf("IVS%s%s", intr, m[1]),
				)
		}

	}

	if rsID.MatchString(item["rsID"]) {
		searchKey = append(searchKey, item["rsID"])
	}
	altKey := strings.Join(searchKey, "\" | \"")
	return gene + " (\"" + altKey + "\")"
}

var (
	//isBLB        = regexp.MustCompile(`B|LB`)
	isClinVarPLP = regexp.MustCompile(`Pathogenic|Likely_pathogenic`)
	//isHgmdDM     = regexp.MustCompile(`DM$|DM\|`)
	isHgmdDMplus = regexp.MustCompile(`DM`)
	//isHgmdDMQ= regexp.MustCompile(`DM\?`)
	isPP3  = regexp.MustCompile(`PP3`)
	isZero = regexp.MustCompile(`^0$|^\.$|^$`)
)

var keys = []string{
	"ExAC HomoAlt Count",
	"PVFD Homo Count",
	"GnomAD HomoAlt Count",
	"1000G EAS AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC EAS AF",
	"ExAC AF",
	"GnomAD EAS AF",
	"GnomAD AF",
	"PVFD AF",
	"Panel AlleleFreq",
}

func tag1(item map[string]string, specVarDb map[string]bool, isTrio bool) string {
	var flag1, flag2 bool
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq <= 0.01 {
		flag1 = true
	}
	if specVarDb[item["MutationName"]] {
		flag1 = true
	}
	if isHgmdDMplus.MatchString(item["HGMD Pred"]) {
		flag1 = true
	}
	if isClinVarPLP.MatchString(item["ClinVar Significance"]) {
		flag1 = true
	}
	if item["遗传相符"] == "相符" {
		if isTrio {
			flag2 = true
		} else {
			inherit := item["ModeInheritance"]
			if isAR.MatchString(inherit) || isXL.MatchString(inherit) || isYL.MatchString(inherit) {
				flag2 = true
			} else if isAD.MatchString(inherit) {
				var flag = true
				for _, key := range keys {
					if !isZero.MatchString(item[key]) {
						flag = false
					}
				}
				if flag {
					flag2 = true
				}
			}
		}
	}
	if flag1 && flag2 {
		return "1"
	}
	return ""
}

func tag2(item map[string]string, specVarDb map[string]bool) string {
	var flag1 bool
	if specVarDb[item["MutationName"]] {
		flag1 = true
	}
	if isHgmdDMplus.MatchString(item["HGMD Pred"]) {
		flag1 = true
	}
	if isClinVarPLP.MatchString(item["ClinVar Significance"]) {
		flag1 = true
	}
	if flag1 {
		return "2"
	}
	return ""
}

func tag3(item map[string]string) string {
	var flag1, flag2 bool
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq <= 0.01 {
		flag1 = true
	}
	if item["烈性突变"] == "是" {
		flag2 = true
	}
	if flag1 && flag2 {
		return "3"
	}
	return ""
}

var tag4Func = map[string]bool{
	"stop-loss": true,
	"cds-del":   true,
	"cds-indel": true,
	"cds-ins":   true,
}

func tag4(item map[string]string) string {
	var flag1, flag2 bool
	frequency := item["frequency"]
	if frequency == "" || frequency == "." {
		frequency = "0"
	}
	freq, e := strconv.ParseFloat(frequency, 32)
	if e != nil {
		log.Printf("%s ParseFloat error:%v", frequency, e)
		freq = 0
	}
	if freq <= 0.01 {
		flag1 = true
	}
	if isPP3.MatchString(item["autoRuleName"]) {
		flag2 = true
	}
	if tag4Func[item["Function"]] && (item["RepeatTag"] == "." || item["RepeatTag"] == "") {
		flag2 = true
	}

	if flag1 && flag2 {
		return "4"
	}
	return ""
}

func tag5(item map[string]string) string {
	if item["SecondaryFinding_Var_致病等级"] != "" {
		return "5"
	}
	return ""
}

//UpdateTags return 筛选标签
func UpdateTags(item map[string]string, specVarDb map[string]bool, isTrio bool) string {
	Tag1 := tag1(item, specVarDb, isTrio)
	Tag2 := tag2(item, specVarDb)
	Tag3 := tag3(item)
	Tag4 := tag4(item)
	Tag5 := tag5(item)
	return strings.Join([]string{Tag1, Tag2, Tag3, Tag4, Tag5}, "")
}

//UpdateFunction fix splice+-20, have not implement
func UpdateFunction(item map[string]string) {
	function := item["Function"]
	if function != "intron" {
		return
	}
}

var floatFormatArray = []string{
	"GnomAD AF",
	//"GnomAD EAS AF",
}

//FloatFormat warp strconv.FormatFloat
func FloatFormat(item map[string]string) {
	for _, key := range floatFormatArray {
		value := item[key]
		if value == "" || value == "." {
			item[key] = ""
			return
		}
		floatValue, e := strconv.ParseFloat(value, 64)
		if e != nil {
			log.Printf("can not ParseFloat:%s[%s]\n", key, value)
		} else {
			item[key] = strconv.FormatFloat(floatValue, 'f', -1, 64)
		}
	}
}

var isBR = regexp.MustCompile(`<br/>`)
var newlineFormatArray = []string{
	"SecondaryFinding_Var_证据项",
	"SecondaryFinding_Var_致病等级",
	"SecondaryFinding_Var_参考文献",
	"SecondaryFinding_Var_Phenotype_OMIM_ID",
	"SecondaryFinding_Var_DiseaseNameEN",
	"SecondaryFinding_Var_DiseaseNameCH",
	"SecondaryFinding_Var_updatetime",
}

//NewlineForamt warp strings.Replace
func NewlineFormat(item map[string]string) {
	for _, key := range newlineFormatArray {
		item[key] = isBR.ReplaceAllString(item[key], "\n")
	}
}

func Format(item map[string]string) {
	FloatFormat(item)
	NewlineFormat(item)
}

//UpdateDisease add disease info to item
func UpdateDisease(gene string, item, gDiseaseDbColumn map[string]string, geneDiseaseDb map[string]map[string]string) {
	// 基因-疾病
	gDiseaseDb := geneDiseaseDb[gene]
	for key, value := range gDiseaseDbColumn {
		item[value] = gDiseaseDb[key]
	}
}

// AFlist default AF list for check
var AFlist = []string{
	"GnomAD EAS AF",
	"GnomAD AF",
	"1000G AF",
	"ESP6500 AF",
	"ExAC EAS AF",
	"ExAC AF",
	"PVFD AF",
	"Panel AlleleFreq",
	"wgs_GnomAD_AF",
}

package anno

// map[string]string update
import (
	"github.com/liserjrqlxue/simple-util"
	"regexp"
	//"github.com/liserjrqlxue/acmg2015"
	"strconv"
)

// Tier1 >1
// LoF 3
var FuncInfo = map[string]int{
	"splice-3":     3,
	"splice-5":     3,
	"init-loss":    3,
	"alt-start":    3,
	"frameshift":   3,
	"nonsense":     3,
	"stop-gain":    3,
	"stop-loss":    2,
	"span":         3,
	"missense":     2,
	"cds-del":      2,
	"cds-indel":    2,
	"cds-ins":      2,
	"splice-10":    2,
	"splice+10":    2,
	"coding-synon": 1,
	"splice-20":    1,
	"splice+20":    1,
}

// regexp
var (
	isHgmd    = regexp.MustCompile("DM")
	isClinvar = regexp.MustCompile("Pathogenic|Likely_pathogenic")
	//newlineReg = regexp.MustCompile(`\n+`)
	isDenovo  = regexp.MustCompile(`^(Hom|Het|Hemi);NA;NA`)
	noProband = regexp.MustCompile(`^NA`)
	isChrAXY  = regexp.MustCompile(`[0-9XY]+$`)
)

// add Tier to item
func AddTier(item map[string]string, stats map[string]int, geneList, specVarDb map[string]bool, isTrio, isWGS, allGene bool, AFlist []string) {
	if isTrio {
		if noProband.MatchString(item["Zygosity"]) {
			stats["noProband"]++
			return
		}
		checkTierTrio(item, stats, geneList, isWGS, allGene, AFlist)
	} else {
		checkTierSingle(item, stats, geneList, isWGS, allGene, AFlist)
	}

	// HGMD or ClinVar
	checkHGMDClinVar(item, stats, AFlist)

	// 特殊位点库
	checkSpecVar(item, stats, specVarDb)

	if item["Tier"] == "Tier1" {
		stats["Tier1"]++
	} else if item["Tier"] == "Tier2" {
		stats["Tier2"]++
	} else if item["Tier"] == "Tier3" {
		stats["Tier3"]++
	} else {
		return
	}
	stats["Retain"]++
	return
}

func checkSpecVar(item map[string]string, stats map[string]int, specVarDb map[string]bool) {
	// 特殊位点库
	if isSpecVar(specVarDb, item["MutationName"]) {
		item["Tier"] = "Tier1"
		stats["SpecVar"]++
	}
}

func checkHGMDClinVar(item map[string]string, stats map[string]int, AFlist []string) {
	if isHgmd.MatchString(item["HGMD Pred"]) || isClinvar.MatchString(item["ClinVar Significance"]) {
		stats["HGMD/ClinVar"]++
		if checkAF(item, AFlist, 0.05) {
			stats["HGMD/ClinVar isAF"]++
			if isChrAXY.MatchString(item["#Chr"]) {
				item["Tier"] = "Tier1"
				stats["HGMD/ClinVar noMT T1"]++
			}
		} else {
			stats["HGMD/ClinVar noAF"]++
			if isChrAXY.MatchString(item["#Chr"]) {
				stats["HGMD/ClinVar noMT T2"]++
				if item["Tier"] != "Tier1" {
					item["Tier"] = "Tier2"
				}
			}
		}
	}
}

func checkTierSingle(item map[string]string, stats map[string]int, geneList map[string]bool, isWGS, allGene bool, AFlist []string) {
	gene := item["Gene Symbol"]
	// Tier
	if item["自动化判断"] != "B" && item["自动化判断"] != "LB" {
		stats["noB/LB"]++
		if checkAF(item, AFlist, 0.01) {
			stats["isAF"]++
			if geneList[gene] || allGene {
				stats["isGene"]++
				if FuncInfo[item["Function"]] > 1 {
					item["Tier"] = "Tier1"
					stats["isFunction"]++
				} else if FuncInfo[item["Function"]] > 0 {
					//pp3,err:=strconv.Atoi(item["PP3"])
					//if err==nil && pp3>0{
					item["Tier"] = "Tier1"
					stats["isFunction"]++
					//}
				} else if isWGS && item["Function"] == "intron" {
					if checkAF(item, []string{"inhouse_AF"}, 0.1) {
						item["Tier"] = "Tier1"
						stats["isFunction"]++
					} else {
						item["Tier"] = "Tier3"
						stats["noFunction"]++
					}
				} else {
					item["Tier"] = "Tier3"
					stats["noFunction"]++
				}
			} else {
				item["Tier"] = "Tier3"
				stats["noGene"]++
			}
		} else {
			item["Tier"] = "Tier3"
			stats["noAF"]++
		}
	}
}

func checkTierTrio(item map[string]string, stats map[string]int, geneList map[string]bool, isWGS, allGene bool, AFlist []string) {
	gene := item["Gene Symbol"]
	// Tier
	if noProband.MatchString(item["Zygosity"]) {
		stats["noProband"]++
		return
	}
	if isDenovo.MatchString(item["Zygosity"]) {
		stats["Denovo"]++
	}
	if item["自动化判断"] != "B" && item["自动化判断"] != "LB" {
		stats["noB/LB"]++
		if isDenovo.MatchString(item["Zygosity"]) {
			stats["isDenovo noB/LB"]++
			if checkAF(item, AFlist, 0.01) {
				stats["low AF"]++
				stats["Denovo AF"]++
				if geneList[gene] || allGene {
					stats["OMIM Gene"]++
					stats["Denovo Gene"]++
					if FuncInfo[item["Function"]] > 1 {
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["Denovo Function"]++
					} else if FuncInfo[item["Function"]] > 0 {
						//pp3,err:=strconv.Atoi(item["PP3"])
						//if err==nil && pp3>0{
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["Denovo Function"]++
					} else if isWGS && item["Function"] == "intron" {
						if checkAF(item, []string{"inhouse_AF"}, 0.1) {
							item["Tier"] = "Tier1"
							stats["Function"]++
							stats["Denovo Function"]++
						} else {
							item["Tier"] = "Tier2"
							stats["noFunction"]++
							stats["Denovo noFunction"]++
						}
					} else {
						item["Tier"] = "Tier2"
						stats["noFunction"]++
						stats["Denovo noFunction"]++
					}
				} else {
					item["Tier"] = "Tier2"
					stats["noB/LB AF noGene"]++
					stats["Denovo noGene"]++
				}
			} else {
				item["Tier"] = "Tier2"
				stats["noB/LB noAF"]++
				stats["Denovo noAF"]++
			}
			if item["Tier"] == "Tier1" {
				stats["Denovo Tier1"]++
			} else {
				stats["Denovo Tier2"]++
			}
		} else {
			stats["noDenovo noB/LB"]++
			if checkAF(item, AFlist, 0.01) {
				stats["low AF"]++
				stats["noDenovo AF"]++
				if geneList[gene] || allGene {
					stats["OMIM Gene"]++
					stats["noDenovo Gene"]++
					if FuncInfo[item["Function"]] > 1 {
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["noDenovo Function"]++
					} else if FuncInfo[item["Function"]] > 0 {
						//pp3,err:=strconv.Atoi(item["PP3"])
						//if err==nil && pp3>0{
						item["Tier"] = "Tier1"
						stats["Function"]++
						stats["noDenovo Function"]++
						//}
					} else if isWGS && item["Function"] == "intron" {
						if checkAF(item, []string{"inhouse_AF"}, 0.1) {
							item["Tier"] = "Tier1"
							stats["Function"]++
							stats["noDenovo Function"]++
						} else {
							item["Tier"] = "Tier3"
							stats["noFunction"]++
							stats["noDenovo noFunction"]++
						}
					} else {
						item["Tier"] = "Tier3"
						stats["noFunction"]++
						stats["noDenovo noFunction"]++
					}
				} else {
					item["Tier"] = "Tier3"
					stats["noB/LB AF noGene"]++
					stats["noDenovo noGene"]++
				}
			} else {
				item["Tier"] = "Tier3"
				stats["noB/LB noAF"]++
				stats["noDenovo noAF"]++
			}
		}
	} else if isDenovo.MatchString(item["Zygosity"]) {
		stats["Denovo B/LB"]++
	}
}

func checkAF(item map[string]string, AFlist []string, threshold float64) bool {
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

func isSpecVar(db map[string]bool, key string) bool {
	if db[key] {
		return db[key]
	} else {
		return false
	}
}

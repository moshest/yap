package transition

import (
	"bufio"
	. "chukuparser/alg/featurevector"
	. "chukuparser/alg/perceptron"
	"chukuparser/util"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	// "sync"
)

const (
	FEATURE_SEPARATOR              = "+" // separates multiple attribute sources
	ATTRIBUTE_SEPARATOR            = "|" // separates attributes in a source
	TEMPLATE_PREFIX                = ":" // output separator
	GENERIC_SEPARATOR              = "|" // output separator
	FEATURE_REQUIREMENTS_SEPARATOR = "," // separates template from requirements
	REQUIREMENTS_SEPARATOR         = ";" // separates multiple requirements
	APPROX_ELEMENTS                = 20
)

var (
	S0R2l, S0Rl       int  = -1, -1
	_Extractor_AllOut bool = true
	_Zpar_Bug_S0R2L   bool = false
)

type FeatureTemplateElement struct {
	Address    []byte
	Offset     int
	Attributes [][]byte

	ConfStr     string
	IsGenerator bool
}

type FeatureTemplate struct {
	Elements                                   []FeatureTemplateElement
	Requirements                               []string
	ID                                         int
	CachedElementIDs                           []int // where to find the feature elements of the template in the cache
	CachedReqIDs                               []int // cached address required to exist for element
	EWord, EPOS, EWPOS, ERel, EMHost, EMSuffix *util.EnumSet
}

type MorphElement struct {
	MorphType      string
	ElementAddress int
}

func (f FeatureTemplate) String() string {
	if _Extractor_AllOut {
		strs := make([]string, len(f.Elements))
		for i, featureElement := range f.Elements {
			strs[i] = featureElement.ConfStr
		}
		return strings.Join(strs, FEATURE_SEPARATOR)
	} else {
		strs := make([]string, len(f.Elements))
		for i, featureElement := range f.Elements {
			strs[i] = featureElement.ConfStr
		}
		retval := make([]string, len(f.Requirements)+1)
		retval[0] = strings.Join(strs, FEATURE_SEPARATOR)
		for j, req := range f.Requirements {
			retval[j+1] = req
		}
		return strings.Join(retval, REQUIREMENTS_SEPARATOR)

	}
}

func (f FeatureTemplate) Format(val interface{}) string {
	return f.FormatWithGenerator(val, f.Elements[0].IsGenerator)
}

func (f FeatureTemplate) FormatWithGenerator(val interface{}, isGenerator bool) string {
	var (
		valueSlice    []interface{}
		valueOneSlice [1]interface{}
		returnSlice   []string
		returnOne     [1]string
	)
	if isGenerator {
		valueSlice = val.([]interface{})
		returnSlice = make([]string, 0, len(valueSlice))
	} else {
		valueOneSlice[0] = val
		valueSlice = valueOneSlice[0:1]
		returnSlice = returnOne[0:0]
	}
	for _, value := range valueSlice {
		if len(f.CachedElementIDs) == 1 {
			switch string(f.Elements[0].Attributes[0]) {
			case "w":
				returnSlice = append(returnSlice, fmt.Sprintf("%v", f.EWord.ValueOf(value.(int))))
			case "p":
				returnSlice = append(returnSlice, fmt.Sprintf("%v", f.EPOS.ValueOf(value.(int))))
			case "h":
				returnSlice = append(returnSlice, fmt.Sprintf("%v", f.EMHost.ValueOf(value.(int))))
			case "s":
				returnSlice = append(returnSlice, fmt.Sprintf("%v", f.EMSuffix.ValueOf(value.(int))))
			case "wp":
				returnSlice = append(returnSlice, fmt.Sprintf("%v", f.EWPOS.ValueOf(value.(int))))
			case "l":
				returnSlice = append(returnSlice, fmt.Sprintf("%d", value.(int)+1))
			default:
				returnSlice = append(returnSlice, fmt.Sprint("%v", value))
			}
		} else {
			retval := make([]string, len(f.CachedElementIDs))
			var sliceVal []interface{}
			switch valueType := value.(type) {
			case [2]interface{}:
				sliceVal = valueType[0:len(valueType)]
			case [3]interface{}:
				sliceVal = valueType[0:len(valueType)]
			case [4]interface{}:
				sliceVal = valueType[0:len(valueType)]
			case [5]interface{}:
				sliceVal = valueType[0:len(valueType)]
			case []interface{}:
				sliceVal = valueType[0:len(valueType)]
			default:
				panic(fmt.Sprintf("Don't know what to do with %v", value))
			}
			var attribNum int
			for _, element := range f.Elements {
				for _, attrib := range element.Attributes {
					curValue := sliceVal[attribNum]
					var (
						resultArray    []string
						resultOneArray [1]string
						valueArray     []interface{}
						valueOneArray  [1]interface{}
					)

					asArray, isArray := curValue.([]interface{})
					if isArray {
						valueArray = asArray
						resultArray = make([]string, 0, len(asArray))
					} else {
						valueOneArray[0] = curValue
						valueArray = valueOneArray[0:1]
						resultArray = resultOneArray[0:0]
					}
					for _, value := range valueArray {
						switch string(attrib) {
						case "w":
							if value == nil {
								resultArray = append(resultArray, "")
							} else {
								resultArray = append(resultArray, fmt.Sprintf("%v", f.EWord.ValueOf(value.(int))))
							}
						case "h":
							if value == nil {
								resultArray = append(resultArray, "")
							} else {
								resultArray = append(resultArray, fmt.Sprintf("%v", f.EMHost.ValueOf(value.(int))))
							}
						case "s":
							if value == nil {
								resultArray = append(resultArray, "")
							} else {
								resultArray = append(resultArray, fmt.Sprintf("%v", f.EMSuffix.ValueOf(value.(int))))
							}
						case "p":
							if value == nil {
								resultArray = append(resultArray, "-NONE-")
							} else {
								resultArray = append(resultArray, fmt.Sprintf("%v", f.EPOS.ValueOf(value.(int))))
							}
						case "wp":
							if value == nil {
								resultArray = append(resultArray, "/-NONE-")
							} else {
								ew := f.EWPOS.ValueOf(value.(int)).([2]string)
								resultArray = append(resultArray, fmt.Sprintf("%s/%s", ew[0], ew[1]))
							}
						case "l":
							log.Println("Printing label")
							log.Println(value)
							if value == nil {
								resultArray = append(resultArray, "-NONE-")
							} else {
								resultArray = append(resultArray, fmt.Sprintf("%d", value.(int)+1))
							}
						case "d":
							if value != nil {
								resultArray = append(resultArray, fmt.Sprintf("%d", value.(int)))
							} else {
								resultArray = append(resultArray, "")
							}
						case "vl", "vr", "vf", "o":
							resultArray = append(resultArray, fmt.Sprintf("%d", value.(int)))
						case "sl", "sr", "sf":
							if value == nil {
								resultArray = append(resultArray, "[ ]")
							}
							if value != nil {
								switch valType := value.(type) {
								case int:
									resultArray = append(resultArray, fmt.Sprintf("[ %v ]", f.ERel.ValueOf(valType)))
								case []int:
									set := valType
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [2]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [3]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [4]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [5]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [6]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [7]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								default:
									resultArray = append(resultArray, fmt.Sprintf("%v", valType))
								}
							}
						case "fp":
							if value == nil {
								resultArray = append(resultArray, "[ ]")
							}
							if value != nil {
								switch valType := value.(type) {
								case int:
									resultArray = append(resultArray, fmt.Sprintf("[ %v ]", f.EPOS.ValueOf(valType)))
								case []int:
									set := valType
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [2]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [3]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [4]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [5]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [6]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								case [7]int:
									set := valType[:]
									tags := make([]string, len(set))
									for i, tag := range set {
										tags[i] = fmt.Sprintf("%v", f.EPOS.ValueOf(tag))
									}
									resultArray = append(resultArray, fmt.Sprintf("[ %s ]", strings.Join(tags, " ")))
								default:
									resultArray = append(resultArray, fmt.Sprintf("%v", valType))
								}
							}
						default:
							panic("Don't know what to do with attribute")
							resultArray = append(resultArray, fmt.Sprint("%v", value))
						}
					}
					retval[attribNum] = fmt.Sprintf("%v", resultArray)
					attribNum++
				}
			}
			returnSlice = append(returnSlice, strings.Join(retval, " "))
		}
	}
	return strings.Join(returnSlice, ",")
}

type GenericExtractor struct {
	FeatureTemplates []FeatureTemplate
	EFeatures        *util.EnumSet

	ElementEnum *util.EnumSet
	AddressEnum *util.EnumSet
	Elements    []FeatureTemplateElement

	Concurrent bool

	Log                                        bool
	EWord, EPOS, EWPOS, ERel, EMHost, EMSuffix *util.EnumSet
}

// Verify GenericExtractor is a FeatureExtractor
var _ FeatureExtractor = &GenericExtractor{}

func (x *GenericExtractor) Init() {
	x.ElementEnum = util.NewEnumSet(APPROX_ELEMENTS)
	x.Elements = make([]FeatureTemplateElement, 0, APPROX_ELEMENTS)
}

func (x *GenericExtractor) Features(instance Instance) []Feature {
	conf, ok := instance.(Configuration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

	features := make([]Feature, len(x.FeatureTemplates))
	{
		// if x.Concurrent {
		// 	featureChan := make(chan interface{})
		// 	wg := new(sync.WaitGroup)
		// 	for i, _ := range x.FeatureTemplates {
		// 		wg.Add(1)
		// 		go func(j int) {
		// 			defer wg.Done()
		// 			valuesArray := make([]interface{}, 0, 5)
		// 			attrArray := make([]interface{}, 0, 5)
		// 			featTemplate := x.FeatureTemplates[j]
		// 			feature, exists := x.GetFeature(conf, featTemplate, valuesArray, attrArray)
		// 			if exists {
		// 				featureChan <- feature
		// 			}
		// 		}(i)
		// 	}
		// 	go func() {
		// 		wg.Wait()
		// 		close(featureChan)
		// 	}()
		// 	for feature := range featureChan {
		// 		features = append(features, Feature(feature))
		// 	}
		// } else {
	}
	if x.Log {
		// log.Println("Generating elements:")
	}
	elementCache := make([]interface{}, len(x.Elements))
	attrArray := make([]interface{}, 0, 5)
	if _Zpar_Bug_S0R2L && (S0R2l < 0 || S0Rl < 0) {
		panic(fmt.Sprintf("Did not set hard coded S0R2l or S0Rl %v", _Zpar_Bug_S0R2L))
	}
	// build element cache
	for i, _ := range x.Elements {
		// log.Println("At template", i, elementTemplate.ConfStr)
		element, exists := x.GetFeatureElement(conf, &x.Elements[i], attrArray[0:0])
		if exists {
			// if x.Log {
			// 	log.Printf("%d %s: %v , isGen = %v\n", i, elementTemplate.ConfStr, element, elementTemplate.IsGenerator)
			// }
			// zpar bug parity
			if _Zpar_Bug_S0R2L && i == S0R2l { // un-documented code in zpar uses S0rl instead of S0r2l (wtf?!)
				// log.Println("Zpar parity")
				elementCache[i] = elementCache[S0Rl]
			} else {
				elementCache[i] = element
			}
			// end zpar bug parity
		} else {
			if x.Log {
				// log.Printf("%d %s: nil\n", i, elementTemplate.ConfStr)
			}
			elementCache[i] = nil
		}
	}
	// if x.Log {
	// 	log.Println("Second template loop:")
	// }

	// for _, elementTemplate := range x.Elements {
	// 	if x.Log {
	// 		log.Println("Template", elementTemplate.ConfStr, "isGen", elementTemplate.IsGenerator)
	// 	}
	// }
	if x.Log {
		// log.Println("Generating features:")
	}
	// generate features
	valuesArray := make([]interface{}, 0, 5)
	var (
		valuesSlice       []interface{}
		hasNilRequirement bool
	)
	for i, template := range x.FeatureTemplates {
		valuesSlice = valuesArray[0:0]
		hasNilRequirement = false
		if x.Log {
			log.Printf("\tTemplate %s; Requirements %v\n", template, template.Requirements)
		}
		for _, reqid := range template.CachedReqIDs {
			if elementCache[reqid] == nil {
				hasNilRequirement = true
				break
			}
		}
		if hasNilRequirement {
			features[i] = nil
		} else {
			(&x.FeatureTemplates[i]).Elements[0].IsGenerator = x.Elements[template.CachedElementIDs[0]].IsGenerator
			if x.Elements[template.CachedElementIDs[0]].IsGenerator {
				if x.Log {
					log.Printf("\t\tIsGenerator")
				}
				generatedElements := elementCache[template.CachedElementIDs[0]].([]interface{})
				fullFeature := make([]interface{}, len(generatedElements))
				// log.Println("\t\tGenerated elements:", generatedElements)
				// log.Println("\t\tCached Elements IDs (0 is generator):", template.CachedElementIDs)
				for j, generatedElement := range generatedElements {
					valuesSlice = valuesSlice[0:0]
					valuesSlice = append(valuesSlice, generatedElement)
					// log.Println("\t\tValues Slice", valuesSlice)
					for _, offset := range template.CachedElementIDs[1:] {
						// valuesSlice = valuesSlice[1:]
						if x.Log {
							log.Printf("\t\t\t(%d,%s): %v", offset, x.Elements[offset].ConfStr, elementCache[offset])
						}
						valuesSlice = append(valuesSlice, elementCache[offset])
					}
					// log.Println("\t\tValues Slice", valuesSlice)
					fullFeature[j] = GetArray(valuesSlice)
				}
				features[i] = fullFeature
				// log.Println("\t\tGenerated", fullFeature)
			} else {
				if x.Log {
					// log.Printf("\t\tIsGenerator false")
				}
				for _, offset := range template.CachedElementIDs {
					if x.Log {
						log.Printf("\t\t(%d,%s): %v", offset, x.Elements[offset].ConfStr, elementCache[offset])
					}
					valuesSlice = append(valuesSlice, elementCache[offset])
				}
				features[i] = GetArray(valuesSlice)
			}
			if x.Log {
				log.Printf("\t\t%s", template.FormatWithGenerator(features[i], x.Elements[template.CachedElementIDs[0]].IsGenerator))
			}
		}
	}
	// valuesArray := make([]interface{}, 0, 5)
	// attrArray := make([]interface{}, 0, 5)
	// for _, tmpl := range x.FeatureTemplates {
	// 	feature, exists := x.GetFeature(conf, tmpl, valuesArray[0:0], attrArray[0:0])
	// 	if exists {
	// 		features = append(features, feature)
	// 	}
	// }
	// }
	return features
}

func (x *GenericExtractor) EstimatedNumberOfFeatures() int {
	return len(x.FeatureTemplates)
}

func (x *GenericExtractor) GetFeature(conf Configuration, template FeatureTemplate, featureValues, attrValues []interface{}) (interface{}, bool) {
	// featureValues := make([]interface{}, 0, len(template.Elements))
	for i, templateElement := range template.Elements {
		featureValues = append(featureValues, nil)
		// check if feature element was already computed
		// cachedValue, cacheExists := x.FeatureResultCache[templateElement.ConfStr]
		cacheExists := false
		if cacheExists {
			// featureValues = append(featureValues, cachedValue)
		} else {
			attrValues = attrValues[0:0]
			elementValue, exists := x.GetFeatureElement(conf, &templateElement, attrValues[0:0])
			if !exists {
				return nil, false
			}
			// x.FeatureResultCache[templateElement.ConfStr] = elementValue
			featureValues[i] = elementValue
		}
	}
	if !x.Concurrent {
		return [3]interface{}{conf.GetLastTransition(), template.ID, GetArray(featureValues)}, true
	} else {
		return GetArray(featureValues), true
	}
}

func (x *GenericExtractor) GetFeatureElement(conf Configuration, templateElement *FeatureTemplateElement, attrValues []interface{}) (interface{}, bool) {
	if x.Log {
		// log.Println(templateElement.ConfStr)
		// log.Println("\tAddress", templateElement.Offset)
	}
	var (
		addresses     []int
		singleAddress [1]int
		resultArray   []interface{}
		singleResult  [1]interface{}
	)
	address, exists, isGenerator := conf.Address([]byte(templateElement.Address), templateElement.Offset)
	if !exists {
		if x.Log {
			// log.Println("\tAddress", templateElement.Offset, "doesnt exist")
		}
		return nil, false
	}
	if x.Log {
		// log.Println("\tAddress", templateElement.Offset, "exists; isGenerator = ", isGenerator)
	}
	// attrValues := make([]interface{}, len(templateElement.Attributes))

	if isGenerator {
		addresses = conf.GenerateAddresses(address, []byte(templateElement.Address))
		resultArray = make([]interface{}, len(addresses))
		templateElement.IsGenerator = true
	} else {
		singleAddress[0] = address
		addresses = singleAddress[0:1]
		resultArray = singleResult[0:1]
	}
	for addressID, generatedAddress := range addresses {
		attrValues = attrValues[0:0]
		for i, attribute := range templateElement.Attributes {
			// if x.Log {
			// 	log.Printf("\t\tAttribute %s\n", attribute)
			// }

			attrValues = append(attrValues, nil)
			attrValue, exists := conf.Attribute(byte(templateElement.Address[0]), generatedAddress, []byte(attribute))
			if !exists {
				// if x.Log {
				// 	log.Printf("\t\tAttribute %s doesnt exist\n", attribute)
				// }
				return nil, false
			}
			// if x.Log {
			// 	log.Printf("\t\tAttribute %s value %v\n", attribute, attrValue)
			// }
			attrValues[i] = attrValue
		}
		resultArray[addressID] = GetArray(attrValues)
	}
	if isGenerator {
		return resultArray, true
	} else {
		return resultArray[0], true
	}
}

func (x *GenericExtractor) ParseFeatureElement(featElementStr string) (*FeatureTemplateElement, error) {
	featElementStrPatchedWP := strings.Replace(featElementStr, "w|p", "wp", -1)
	elementParts := strings.Split(featElementStrPatchedWP, ATTRIBUTE_SEPARATOR)

	if len(elementParts) < 2 {
		return nil, errors.New("Not enough parts for element " + featElementStr)
	}

	// TODO: add validation to element parts
	element := new(FeatureTemplateElement)

	element.ConfStr = featElementStrPatchedWP
	element.Address = []byte(elementParts[0])
	// TODO fix to get more than one digit of offset
	parsedOffset, err := strconv.ParseInt(string(element.Address[1]), 10, 0)
	element.Offset = int(parsedOffset)
	if err != nil {
		panic("Error parsing feature element " + featElementStr + " " + err.Error())
	}
	element.Attributes = make([][]byte, len(elementParts)-1)

	for i, elementStr := range elementParts[1:] {
		element.Attributes[i] = []byte(elementStr)
	}
	return element, nil
}

func (x *GenericExtractor) ParseMorphConfiguration(morphTemplateStr string) *MorphElement {
	parts := strings.Split(morphTemplateStr, ATTRIBUTE_SEPARATOR)
	tmpl := new(MorphElement)
	tmpl.MorphType = parts[0][1:] // remove 'M' from morphological feature
	if len(parts) > 1 {
		parsedOffset, err := strconv.ParseInt(parts[1], 10, 0)
		if err != nil {
			panic("Error parsing morph feature element " + morphTemplateStr + " " + err.Error())
		}
		tmpl.ElementAddress = int(parsedOffset) - 1
	} else {
		tmpl.ElementAddress = 0
	}
	return tmpl
}

func (x *GenericExtractor) ParseFeatureTemplate(featTemplateStr string, requirements string) (*FeatureTemplate, error) {
	// remove any spaces
	featTemplateStr = strings.Replace(featTemplateStr, " ", "", -1)
	features := strings.Split(featTemplateStr, FEATURE_SEPARATOR)
	featureTemplate := make([]FeatureTemplateElement, len(features))

	for i, featElementStr := range features {
		// TODO: morph template is a hack, should be more generic
		// if featElementStr[0] == 'M' { // element is a morphological template
		// 	morphElement := x.ParseMorphConfiguration(featElementStr)
		// 	newMorphElement := new(FeatureTemplateElement)
		// 	refElement := featureTemplate[morphElement.ElementAddress]
		// 	newMorphElement.Address = refElement.Address
		// 	newMorphElement.ConfStr = featElementStr
		// 	newMorphElement.IsGenerator = false
		// 	newMorphElement.Attributes = make([][]byte, 1)
		// 	newMorphElement.Attributes[0] = []byte(morphElement.MorphType)
		// } else {
		parsedElement, err := x.ParseFeatureElement(featElementStr)
		if err != nil {
			return nil, err
		}
		featureTemplate[i] = *parsedElement
		// }
	}
	reqArr := strings.Split(requirements, REQUIREMENTS_SEPARATOR)
	return &FeatureTemplate{Elements: featureTemplate, Requirements: reqArr,
		EWord: x.EWord, EPOS: x.EPOS, EWPOS: x.EWPOS, ERel: x.ERel, EMHost: x.EMHost, EMSuffix: x.EMSuffix}, nil
}

func (x *GenericExtractor) UpdateFeatureElementCache(feat *FeatureTemplate) {
	// log.Println("Update cache for", feat)
	feat.CachedElementIDs = make([]int, 0, len(feat.Elements))
	var (
		elementId int
		isNew     bool
	)
	for _, element := range feat.Elements {
		// log.Println("\tElement", element.ConfStr)
		for _, attr := range element.Attributes {
			fullConfStr := new(string)
			*fullConfStr = string(element.Address) + "|" + string(attr)
			// log.Println("\t\tAttribute", *fullConfStr)
			elementId, isNew = x.ElementEnum.Add(*fullConfStr)
			if isNew {
				// zpar parity
				if *fullConfStr == "S0r2|l" {
					S0R2l = elementId
				}
				if *fullConfStr == "S0r|l" {
					S0Rl = elementId
				}
				// end zpar parity
				fullElement := new(FeatureTemplateElement)
				fullElement.Address = element.Address
				fullElement.Offset = element.Offset
				fullElement.Attributes = make([][]byte, 1)
				fullElement.Attributes[0] = attr
				fullElement.ConfStr = *fullConfStr
				x.Elements = append(x.Elements, *fullElement)
				// log.Println("\t\tGenerated", fullElement.ConfStr)
			}
			// log.Println("\t\tID:", elementId)
			feat.CachedElementIDs = append(feat.CachedElementIDs, elementId)
		}
	}
	feat.CachedReqIDs = make([]int, len(feat.Requirements))
	for i, req := range feat.Requirements {
		reqid, exists := x.ElementEnum.IndexOf(req)
		if !exists {
			panic(fmt.Sprintf("Can't find requirement element %s for features %s", req, feat))
		}
		feat.CachedReqIDs[i] = reqid
	}
}

func (x *GenericExtractor) LoadFeature(featTemplateStr string, requirements string) error {
	template, err := x.ParseFeatureTemplate(featTemplateStr, requirements)
	if err != nil {
		return err
	}
	x.UpdateFeatureElementCache(template)
	template.ID, _ = x.EFeatures.Add(featTemplateStr)
	x.FeatureTemplates = append(x.FeatureTemplates, *template)
	return nil
}

func (x *GenericExtractor) LoadFeatures(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	// scan lines, lines beginning with # are ommitted
	for scanner.Scan() {
		line := scanner.Text()
		// skip blank and comment lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		// parse feature
		if err := x.LoadFeature(line, ""); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (x *GenericExtractor) LoadFeatureSetup(setup *FeatureSetup) {
	// load morphological templates for feature group combinations
	morphGroups := make(map[string]int)
	for i, morphGroup := range setup.MorphTemplates {
		morphGroups[morphGroup.Group] = i
	}

	// load feature groups
	var (
		featurePair       []string
		morphCombinations []string
		morphId           int
		exists            bool
		morphAddedFeature string
	)
	for _, group := range setup.FeatureGroups {
		log.Println("Loading feature group", group.Group)
		morphId, exists = morphGroups[group.Group]
		if exists {
			morphCombinations = setup.MorphTemplates[morphId].Combinations
			log.Println(" with morph combinations ", fmt.Sprintf("%v", morphCombinations))
		} else {
			morphCombinations = nil
		}
		for _, featureConfig := range group.Features {
			// a feature pair is a feature with it's requirement:
			// e.g. S0p,S0w: feature is S0p, requires S0w
			featurePair = strings.Split(featureConfig, FEATURE_REQUIREMENTS_SEPARATOR)
			log.Println("\tLoading feature", featurePair[0])
			if err := x.LoadFeature(featurePair[0], featurePair[1]); err != nil {
				log.Fatalln("Failed to load feature", err.Error())
			}
			if morphCombinations != nil {
				for _, morphTmpl := range morphCombinations {
					morphAddedFeature = fmt.Sprintf("%s%s%s", featurePair[0], FEATURE_SEPARATOR, morphTmpl)
					log.Println("\t generating with morph ", morphAddedFeature)
					if err := x.LoadFeature(morphAddedFeature, featurePair[1]); err != nil {
						log.Fatalln("Failed to load morph feature", err.Error())
					}
				}
			}
		}
	}
}

func GetArray(input []interface{}) interface{} {
	switch len(input) {
	case 0:
		return nil
	case 1:
		return input[0]
	case 2:
		return [2]interface{}{input[0], input[1]}
	case 3:
		return [3]interface{}{input[0], input[1], input[2]}
	case 4:
		return [4]interface{}{input[0], input[1], input[2], input[3]}
	case 5:
		return [5]interface{}{input[0], input[1], input[2], input[3], input[4]}
	case 6:
		return [6]interface{}{input[0], input[1], input[2], input[3], input[4], input[5]}
	default:
		result := make([]string, len(input))
		for i, val := range input {
			result[i] = fmt.Sprintf("%v", val)
		}
		return strings.Join(result, GENERIC_SEPARATOR)
	}
}

func GetArrayInt(input []int) interface{} {
	switch len(input) {
	case 0:
		return nil
	case 1:
		return input[0]
	case 2:
		return [2]int{input[0], input[1]}
	case 3:
		return [3]int{input[0], input[1], input[2]}
	case 4:
		return [4]int{input[0], input[1], input[2], input[3]}
	case 5:
		return [5]int{input[0], input[1], input[2], input[3], input[4]}
	case 6:
		return [6]int{input[0], input[1], input[2], input[3], input[4], input[5]}
	default:
		result := make([]string, len(input))
		for i, val := range input {
			result[i] = fmt.Sprintf("%v", val)
		}
		return strings.Join(result, GENERIC_SEPARATOR)
	}
}
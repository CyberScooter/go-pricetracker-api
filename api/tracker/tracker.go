
package tracker

import (
	"fmt"
	"time"
	"regexp"
	"strings"
	"strconv"
	"tracker-api-backend-go/api/structs"

	"github.com/tebeka/selenium"
	"github.com/serge1peshcoff/selenium-go-conditions"
)


/**
* ? fix issue with random page being loaded when searching for product, it stops bot from progressing
* *Temporary fix has been made, if product is still searching for more than a second then respond with not found
* TODO: add headless and icognito capabilities
* TODO: optimise it so its faster

*/

func Track(item string) []structs.Product {
// func main(){
	// FireFox driver without specific version
	caps := selenium.Capabilities{"browserName": "chrome"}

	// chromeCaps := chrome.Capabilities{
	// 	Path:  "",
	// 	Args: []string{
	// 		"--headless", // <<<
	// 		"--no-sandbox",
	// 	},
	// }
	// caps.AddChrome(chromeCaps)

	wd, _ := selenium.NewRemote(caps, "")
	defer wd.Quit()

	items := make([]structs.Product, 0)

	data := item

	// Get simple playground interface
	wd.Get("http://google.com/search?q=" +  data + "&tbm=shop&tbs=vw:g")

	// dont need it
	// wd.SwitchFrame(0)
	// fr, _ := wd.FindElement(selenium.ByID, "introAgreeButton")
	// fr.Click()

	//"//*[text()='Get started free']"
	//"//div[@class='C1iIFb IHk3ob']"


	if err := wd.WaitWithTimeout(conditions.ElementIsLocated(selenium.ByXPATH, "//a[@class='a3H7pd']"), 1*time.Second); err == nil {

		topResult, error := wd.FindElement(selenium.ByXPATH, "//a[@class='a3H7pd']")
		topResult.Click()
		if error != nil {
			panic(error)
		}

		topResult.Click()

		// waiting till element is located before continuing
		if err := wd.Wait(conditions.ElementIsLocated(selenium.ByID, "sh-osd__online-sellers-cont")); err == nil{
			result, _ := wd.FindElement(selenium.ByID, "sh-osd__online-sellers-cont")
			allResult, _ := result.Text()

			// result, _ := wd.FindElement(selenium.ByID, "sh-osd__online-sellers-cont")

			regexPercentagePostive, _ := regexp.Compile("[0-9]+% positive([(]([[0-9]+[,][0-9]+[,]*[0-9]*]*|[0-9]*)[)](\n.*?)*)")
			regexPercentageNegative, _ := regexp.Compile("[0-9]+% negative([(]([[0-9]+[,][0-9]+[,]*[0-9]*]*|[0-9]*)[)](\n.*?)*)")
			regexPricePound, _ := regexp.Compile("£([0-9]+).([0-9]+)")
			regexPriceDollar, _ := regexp.Compile("$([0-9]+).([0-9]+)")
			regexPriceEuro, _ := regexp.Compile("€([0-9]+).([0-9]+)")
			regexFreeDelivery, _ := regexp.Compile("Free delivery")
			
			s := strings.Split(allResult, "Visit site")

			for i := range s {
				s[i] = strings.TrimSpace(s[i])
			}


			// TODO ADD LINKS FOR EACH PRODUCT
			linksElements, _ := wd.FindElements(selenium.ByXPATH, "//a[@class='sh-osd__seller-link shntl']")
			links := []string{}

			for i := range linksElements {
				link, _ := linksElements[i].GetAttribute("href")
				links = append(links, link)
			}

			company, review, delivery, price, totalPrice, URL, currency := "", "", "", 0.00, 0.00, "", ""

			for i, v := range s {
				description := strings.Split(v, "\n")

				tempPrice1 := 0.00
				tempPrice2 := 0.00

				if(i <= len(links) - 1){
					URL = links[i]
				}

				for i, k := range description {

					if i == 0 {
						company = k
					}else if regexPercentagePostive.MatchString(k) || regexPercentageNegative.MatchString(k){
						review = k
					}else if regexPricePound.MatchString(k) || regexPriceEuro.MatchString(k) || regexPriceDollar.MatchString(k){
						if len(k) >= 8{
							f, _ := strconv.ParseFloat(k[2:8], 64)
							if tempPrice1 == 0.00 {
								tempPrice1 = f
							}else {
								tempPrice2 = f
							}
						}else if len(k) < 8 {
							f, _ := strconv.ParseFloat(k[2:], 64)
							if tempPrice1 == 0.00 {
								tempPrice1 = f
							}else {
								tempPrice2 = f
							}
						}
						

						if string([]rune(k)[0]) == "£" {
							currency = "Pound"
						} else if string([]rune(k)[0]) == "$" {
							currency = "Dollar"
						} else if string([]rune(k)[0]) == "€" {
							currency = "Euro"
						}
						
						// if tempPrice1 == 0.00 {
						// 	tempPrice1 = f
						// }else {
						// 	tempPrice2 = f
						// }
					}else if regexFreeDelivery.MatchString(k){
						delivery = k
					}
					
					if i == (len(description) - 1) { 
						if tempPrice1 > tempPrice2 {
							totalPrice, price = tempPrice1, tempPrice2
							r := fmt.Sprintf("%.2f", totalPrice - price)
							delivery = r
						}else if tempPrice1 < tempPrice2 {
							totalPrice, price = tempPrice2, tempPrice1
							r := fmt.Sprintf("%.2f", totalPrice - price)
							delivery = r
						} else if tempPrice1 == 0.00 && tempPrice2 == 0.00{
							break
						}else if tempPrice1 == tempPrice2  {
							totalPrice, price = tempPrice1, tempPrice2
						} 

						items = append(items, structs.Product{Found: true, Company: company, Currency: currency, Review: review, Delivery: delivery, Price: price, TotalPrice: totalPrice, URL: URL})
						company, review, delivery, price, totalPrice, URL, tempPrice1, tempPrice2, currency = "", "", "", 0.00, 0.00, "", 0.00, 0.00, ""
						break
				
					}

				}
			}

		// fmt.Println(items)

		}


	} else {
		return []structs.Product{structs.Product{Found: false, Company: "", Currency: "", Review: "", Delivery: "", Price: 0.00, TotalPrice: 0.00, URL: ""}}
	}
	return items



    // regexPercentagePostive := "/[0-9]+% positive([(]([[0-9]+[,][0-9]+[,]*[0-9]*]*|[0-9]*)[)](\n.*?)*)/gi"
    // regexPercentageNegative := "/[0-9]+% negative([(]([[0-9]+[,][0-9]+[,]*[0-9]*]*|[0-9]*)[)](\n.*?)*)/gi"
    // regexPrice := "/£[0-9]+.([0-9](\n.*?)*)+/gi"
    // regexFreeDelivery := "/Free delivery+/gi"

	

	// // Enter code in textarea
	// elem, _ := wd.FindElement(selenium.ByCSSSelector, "#code")
	// elem.Clear()
	// // elem.SendKeys(code)

	// // Click the run button
	// btn, _ := wd.FindElement(selenium.ByCSSSelector, "#run")
	// btn.Click()

	// // Get the result
	// div, _ := wd.FindElement(selenium.ByCSSSelector, "#output")

	// output := ""
	// // Wait for run to finish
	// for {
	// 	output, _ = div.Text()
	// 	if output != "Waiting for remote server..." {
	// 		break
	// 	}
	// 	time.Sleep(time.Millisecond * 100)
	// }

	// fmt.Printf("Got: %s\n", output)
}